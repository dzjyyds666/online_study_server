package core

import (
	"class/api/lua"
	"common/proto"
	"common/rpc/client"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type ClassServer struct {
	ctx           context.Context
	classDB       *redis.Client
	chapterServer *ChapterServer // 章节服务
	taskServer    *TaskServer    // 作业
}

func NewClassServer(ctx context.Context, dsClient *redis.Client) *ClassServer {
	return &ClassServer{
		ctx:           ctx,
		classDB:       dsClient,
		chapterServer: NewChapterServer(ctx, dsClient),
	}
}

func (cls *ClassServer) CreateClass(ctx context.Context, info *Class) error {
	// 使用lua脚本创建文件夹
	err := cls.classDB.Eval(ctx, lua.CreateClassScript, []string{
		BuildClassInfo(*info.Cid),
		BuildTeacherClassList(*info.Teacher),
		BuildAllClassList(),
	}, info.CreateTs, info.Marshal(), *info.Cid).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("CreateClass|Create Class Error|%v", err)
		return err
	}
	return nil
}

func (cls *ClassServer) RecoverClass(ctx context.Context, cid string) error {
	// 恢复课程
	result, err := cls.classDB.Get(ctx, BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|RecoverClass|GetClassInfoError|%v", err)
		return err
	}

	var class *Class
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|RecoverClass|UnmarshalClassInfoError|%v", err)
		return err
	}
	// 执行恢复操作
	err = cls.classDB.Eval(ctx, lua.RecoverClass, []string{
		BuildTeacherClassList(*class.Teacher),
		BuildTeacherClassDeletedList(*class.Teacher),
		BuildAllClassList(),
		BuildClassDeletedList(),
	}, cid, class.CreateTs).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|RecoverClass|RecoverClassError|%v", err)
	}

	// 修改课程的删除状态
	class.WithDeleted(false)
	err = cls.classDB.Set(ctx, BuildClassInfo(cid), class.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|RecoverClass|UpdateClassInfoError|%v", err)
		return err
	}
	return nil
}

// 把课程移动到回收站
func (cls *ClassServer) MoveClassToTrash(ctx context.Context, cid string) error {
	// 修改课程的删除状态
	result, err := cls.classDB.Get(ctx, BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|MoveClassToTrash|GetClassInfoError|%v", err)
		return err
	}

	var class *Class
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|MoveClassToTrash|UnmarshalClassInfoError|%v", err)
		return err
	}

	err = cls.classDB.Eval(ctx, lua.MoveClassToTrash, []string{
		BuildTeacherClassList(*class.Teacher),
		BuildTeacherClassDeletedList(*class.Teacher),
		BuildAllClassList(),
		BuildClassDeletedList(),
	}, cid, class.CreateTs).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|MoveClassToTrash|UpdateClassInfoError|%v", err)
		return err
	}
	class.WithDeleted(true)
	err = cls.classDB.Set(ctx, BuildClassInfo(cid), class.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|MoveClassToTrash|UpdateClassInfoError|%v", err)
		return err
	}
	return nil
}

// 更新课程信息
func (cls *ClassServer) UpdateClass(ctx context.Context, info *Class) error {
	// 查询原始课程信息
	result, err := cls.classDB.Get(ctx, BuildClassInfo(*info.Cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|UpdateClass|GetClassInfoError|%v", err)
		return err
	}

	class, err := UnmarshalToClass([]byte(result))
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|UpdateClass|UnmarshalClassInfoError|%v", err)
		return err
	}

	cls.updateClassInfo(class, info)

	err = cls.classDB.Set(ctx, BuildClassInfo(*info.Cid), class.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|UpdateClass|UpdateClassInfoError|%v", err)
		return err
	}

	return nil

}

func (cls *ClassServer) QueryClassInfo(ctx context.Context, cid string) (*Class, error) {
	result, err := cls.classDB.Get(ctx, BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryClassInfo|GetClassInfoError|%v", err)
		return nil, err
	}

	var class Class
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryClassInfo|UnmarshalClassInfoError|%v", err)
		return nil, err
	}
	logx.GetLogger("study").Infof("ClassServer|QueryClassInfo|QueryClassInfoSuccess|%v", common.ToStringWithoutError(class))
	return &class, nil
}

func (cls *ClassServer) CreateChapter(ctx context.Context, info *Chapter) error {

	// 把章节信息保存到课程章节下面
	err := cls.classDB.ZAdd(ctx, BuildClassChapterList(*info.SourceId), redis.Z{
		Member: info.Chid,
		Score:  float64(time.Now().Unix()),
	}).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("ChapterServer|CreateChapter|AddChapterToClassError|%v", err)
		return err
	}

	// 创建章节
	err = cls.chapterServer.CreateChapter(ctx, info)
	if err != nil {
		logx.GetLogger("study").Errorf("CreateChapter|Create Chapter Error|%v", err)
		return err
	}

	return nil
}

func (cls *ClassServer) UpdateChapter(ctx context.Context, info *Chapter) error {
	return cls.chapterServer.UpdateChapter(ctx, info)
}

func (cls *ClassServer) DeleteChapter(ctx context.Context, chid string) error {
	chapter, err := cls.chapterServer.QueryChapterInfo(ctx, chid)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteChapter|QueryChapterInfoError|%v", err)
		return err
	}
	// 先从课程章节列表中删除
	err = cls.classDB.ZRem(ctx, BuildClassChapterList(*chapter.SourceId), chapter.Chid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteChapter|DeleteChapterFromClassError|%v", err)
		return err
	}

	err = cls.chapterServer.DeleteChapter(ctx, *chapter.Chid)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteChapter|DeleteChapterError|%v", err)
		return err
	}
	return nil
}

func (cls *ClassServer) QueryChapterList(ctx context.Context, cid string) ([]*Chapter, error) {
	chids, err := cls.classDB.ZRange(ctx, BuildClassChapterList(cid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryChapterList|QueryChapterListError|%v", err)
		return nil, err
	}

	list := make([]*Chapter, 0, len(chids))

	for _, chid := range chids {
		info, err := cls.chapterServer.QueryChapterInfo(ctx, chid)
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|QueryChapterList|QueryChapterInfoError|%v", err)
			return nil, err
		}
		list = append(list, info)
	}
	return list, nil
}

func (cls *ClassServer) CreateResource(ctx context.Context, resource *Resource) error {
	return cls.chapterServer.CreateResource(ctx, resource)
}

func (cls *ClassServer) UpdateResource(ctx context.Context, resource *Resource) (*Resource, error) {
	return cls.chapterServer.UpdateResource(ctx, resource)
}

func (cls *ClassServer) DeleteResource(ctx context.Context, fid string, chid string) error {
	return cls.chapterServer.DeleteResource(ctx, fid, chid)
}

func (cls *ClassServer) updateClassInfo(oldClass, newClass *Class) *Class {
	if newClass.Cover != nil && len(*newClass.Cover) > 0 {
		oldClass.WithCover(*newClass.Cover)
	}
	if newClass.ClassName != nil && len(*newClass.ClassName) > 0 {
		oldClass.WithClassName(*newClass.ClassName)
	}
	if newClass.ClassDesc != nil && len(*newClass.ClassDesc) > 0 {
		oldClass.WithClassDesc(*newClass.ClassDesc)
	}

	if newClass.ClassType != nil && len(*newClass.ClassType) > 0 {
		oldClass.WithClassType(*newClass.ClassType)
	}

	if newClass.Archive != nil {
		oldClass.WithArchive(*newClass.Archive)
	}

	if newClass.Deleted != nil {
		oldClass.WithDeleted(*newClass.Deleted)
	}

	if newClass.ClassScore != nil && len(*newClass.ClassScore) > 0 {
		oldClass.WithClassScore(*newClass.ClassScore)
	}

	if newClass.ClassTime != nil && len(*newClass.ClassTime) > 0 {
		oldClass.WithClassTime(*newClass.ClassTime)
	}

	if newClass.ClassCollege != nil && len(*newClass.ClassCollege) > 0 {
		oldClass.WithClassCollege(*newClass.ClassCollege)
	}

	if newClass.ClassSchoolTerm != nil && len(*newClass.ClassSchoolTerm) > 0 {
		oldClass.WithClassSchoolTerm(*newClass.ClassSchoolTerm)
	}

	if newClass.ClassOutline != nil && len(*newClass.ClassOutline) > 0 {
		oldClass.WithClassOutline(*newClass.ClassOutline)
	}
	logx.GetLogger("study").Infof("ClassServer|UpdateClass|UpdateClassInfoSuccess|%s", common.ToStringWithoutError(*oldClass))
	return oldClass
}

func (cls *ClassServer) QueryClassList(ctx context.Context, uid string) ([]*Class, error) {
	// 查询用户的课程列表
	cids, err := cls.classDB.ZRange(ctx, BuildTeacherClassList(uid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryClassList|QueryClassListError|%v", err)
		return nil, err
	}
	list := make([]*Class, 0, len(cids))
	for _, id := range cids {
		info, err := cls.QueryClassInfo(ctx, id)
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|QueryClassList|QueryClassInfoError|%v", err)
			return nil, err
		}
		list = append(list, info)
	}
	return list, nil
}

func (cls *ClassServer) QueryTeacherDeletedClassList(ctx context.Context, uid string) ([]*Class, error) {
	cids, err := cls.classDB.ZRange(ctx, BuildTeacherClassDeletedList(uid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryTeacherDeletedClassList|QueryTeacherDeletedClassListError|%v", err)
		return nil, err
	}

	list := make([]*Class, len(cids))

	for _, cid := range cids {
		info, err := cls.QueryClassInfo(ctx, cid)
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|QueryTeacherDeletedClassList|QueryTeacherDeletedClassListError|%v", err)
			return nil, err
		}
		list = append(list, info)
	}
	return list, nil
}

func (cls *ClassServer) DeleteClassFromTrash(ctx context.Context, cid string) error {
	// 先查询课程的信息
	class, err := cls.QueryClassInfo(ctx, cid)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteClassFromTrash|QueryClassInfoError|%v", err)
		return err
	}

	if !class.IsDeleted() {
		logx.GetLogger("study").Errorf("ClassServer|DeleteClassFromTrash|ClassNotDeleted|%v", err)
		return errors.New("class not deleted")
	}

	err = cls.classDB.Eval(ctx, lua.DeleteClass, []string{
		BuildTeacherClassDeletedList(*class.Teacher),
		BuildClassDeletedList(),
		BuildClassInfo(cid),
	}, cid).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteClassFromTrash|DeleteClassFromTrashError|%v", err)
		return err
	}

	return nil
}

func (cls *ClassServer) CopyClass(ctx context.Context, cid string) (*Class, error) {
	// 上往下复制
	result, err := cls.classDB.Get(ctx, BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CopyClass|GetClassInfoError|%v", err)
		return nil, err
	}

	var class *Class
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CopyClass|UnmarshalClassInfoError|%v", err)
		return nil, err
	}

	// 生成新的课程
	newCid := NewClassId(8)
	class.WithCid(newCid)

	// 重新写入redis中
	err = cls.classDB.Set(ctx, BuildClassInfo(newCid), class.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CopyClass|SetClassInfoError|%v", err)
		return nil, err
	}

	// 复制章节，先查询到原本的章节列表
	chapterList, err := cls.QueryChapterList(ctx, cid)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CopyClass|QueryChapterListError|%v", err)
		return nil, err
	}

	for _, chapter := range chapterList {
		newchid := NewChapterId(8)
		var chInfo Chapter
		chInfo.WithChid(newchid)
		chInfo.WithSourceId(newCid)
		chInfo.WithChapterName(*chapter.ChapterName)
		err = cls.CreateChapter(ctx, &chInfo)
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|CopyClass|CreateChapterError|%v", err)
			break
		}
		// 添加到课程对应的章节下面
		err = cls.classDB.ZAdd(ctx, BuildClassChapterList(newCid), redis.Z{
			Member: newchid,
			Score:  float64(time.Now().Unix()),
		}).Err()
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|CopyClass|AddChapterToClassError|%v", err)
			break
		}
		// 复制资源
		for _, resource := range chapter.ResourceList {
			rid := NewFid()
			var reInfo Resource
			reInfo.WithChid(newchid).
				WithFid(rid).
				WithDownloadable(false).
				WithPublished(false)
			err := cls.CreateResource(ctx, &reInfo)
			if err != nil {
				logx.GetLogger("study").Errorf("ClassServer|CopyClass|CreateResourceError|%v", err)
				break
			}
			cos := client.GetCosRpcClient(ctx)
			_, err = cos.CopyObject(ctx, &proto.CopyObjectRequest{
				SrcFid: *resource.Fid,
				DstFid: rid,
			})
			if err != nil {
				logx.GetLogger("study").Errorf("ClassServer|CopyClass|CopyObjectError|%v", err)
				return nil, err
			}

			// 把资源添加到章节列表下
			err = cls.classDB.ZAdd(ctx, BuildChapterResourceList(newchid), redis.Z{
				Member: rid,
				Score:  float64(time.Now().Unix()),
			}).Err()

			if err != nil {
				logx.GetLogger("study").Errorf("ClassServer|CopyClass|AddResourceToChapterError|%v", err)
				break
			}
			chInfo.ResourceList = append(chInfo.ResourceList, reInfo)
		}
		class.ChapterList = append(class.ChapterList, chInfo)
	}
	return class, nil
}

func (cls *ClassServer) ImportStudentFromExcel(ctx context.Context, filename, cid string, r io.Reader) ([]string, error) {
	userClient := client.GetUserRpcClient(ctx)
	stream, err := userClient.BatchAddStudentToClass(ctx)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|ImportStudentFromExcel|BatchRegisterError|%v", err)
		return nil, err
	}
	buf := make([]byte, 1024*32)
	firstChunk := true

	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|ImportStudentFromExcel|ReadError|%v", err)
			return nil, err
		}
		chunk := &proto.FileChunk{
			Content: buf[:n],
		}

		if firstChunk {
			chunk.Filename = filename
			chunk.Cid = cid
			firstChunk = false
		}
		if err = stream.Send(chunk); err != nil {
			logx.GetLogger("study").Errorf("ClassServer|ImportStudentFromExcel|SendError|%v", err)
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|ImportStudentFromExcel|CloseAndRecvError|%v", err)
		return nil, err
	}
	return resp.Uids, nil
}

func (cls *ClassServer) UploadClassCover(ctx context.Context, md5, fileType, dirId string, open io.Reader) (string, error) {
	// 调用cos上传文件
	cosClient := client.GetCosRpcClient(ctx)
	stream, err := cosClient.UploadClassCover(ctx)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|UploadClassCover|UploadClassCoverError|%v", err)
		return "", err
	}
	buf := make([]byte, 1024*1024)
	firstChunk := true
	for {
		n, err := open.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|UploadClassCover|ReadError|%v", err)
			return "", err
		}

		chunk := &proto.UploadClassCoverReq{
			Content: buf[:n],
		}
		if firstChunk {
			chunk.Md5 = md5
			chunk.FileType = fileType
			chunk.DirectoryId = dirId
			firstChunk = false
		}
		if err := stream.Send(chunk); err != nil {
			logx.GetLogger("study").Errorf("ClassServer|UploadClassCover|SendError|%v", err)
			return "", err
		}
	}

	recv, err := stream.CloseAndRecv()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|UploadClassCover|CloseAndRecvError|%v", err)
		return "", err
	}
	return recv.Fid, nil
}

func (cls *ClassServer) CreateTask(ctx context.Context, task *Task) error {
	id := NewTaskId(8)
	task.WithId(id)

	err := cls.taskServer.CreateTask(ctx, task)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CreateTask|CreateTaskError|%v", err)
		return err
	}
	err = cls.classDB.ZAdd(ctx, BuildClassTaskList(task.Cid), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: task.TaskId,
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|CreateTask|AddTaskToClassError|%v", err)
		return nil
	}
	return nil
}

func (cls *ClassServer) ListTask(ctx context.Context, list *ListTask) error {
	zrangeBy := &redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(math.MaxInt64, 10),
		Count:  list.Limit,
		Offset: 0,
	}
	if len(list.ReferId) > 0 {
		score, err := cls.classDB.ZScore(ctx, BuildClassTaskList(list.Cid), list.ReferId).Result()
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|ListTask|GetReferIdScoreError|%v", err)
			return err
		}
		zrangeBy.Min = "(" + strconv.FormatInt(int64(score), 10)
	}

	taskIds, err := cls.classDB.ZRangeByScore(ctx, BuildClassTaskList(list.Cid), zrangeBy).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|ListTask|GetTaskIdListError|%v", err)
		return err
	}
	for _, taskId := range taskIds {
		task, err := cls.taskServer.QueryTaskInfo(ctx, taskId)
		if err != nil {
			logx.GetLogger("study").Errorf("ClassServer|ListTask|GetTaskInfoError|%v", err)
			return err
		}
		list.Tasks = append(list.Tasks, task)
	}
	return nil
}

func (cls *ClassServer) DeleteTask(ctx context.Context, tid string) (*Task, error) {
	info, err := cls.taskServer.QueryTaskInfo(ctx, tid)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|DeleteTask|GetTaskInfoError|%v", err)
		return nil, err
	}

	err = cls.taskServer.DeleteTask(ctx, tid)
	if err != nil {
		logx.GetLogger().Errorf("ClassServer|DeleteTask|DeleteTaskError|%v", err)
		return nil, err
	}
	// 移除任务列表中的索引
	err = cls.classDB.ZRem(ctx, BuildClassTaskList(info.Cid), tid).Err()
	if err != nil {
		logx.GetLogger().Errorf("ClassServer|DeleteTask|RemoveTaskFromClassError|%v", err)
		return nil, err
	}
	return info, nil
}

func (cls *ClassServer) QueryResourceList(ctx context.Context, list *ResourceList) error {
	return cls.chapterServer.QueryResourceList(ctx, list)
}
