package core

import (
	"class/api/http/internal/lua"
	"context"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

type ClassServer struct {
	ctx           context.Context
	classDB       *redis.Client
	chapterServer *ChapterServer // 章节服务
}

func NewClassServer(ctx context.Context, dsClient *redis.Client) *ClassServer {
	return &ClassServer{
		ctx:     ctx,
		classDB: dsClient,
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

	class, err := ClassUnmarshal([]byte(result))
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

	var class *Class
	err = json.Unmarshal([]byte(result), class)
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryClassInfo|UnmarshalClassInfoError|%v", err)
		return nil, err
	}

	return class, nil
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

func (cls *ClassServer) DeleteChapter(ctx context.Context, chapter *Chapter) error {
	// 先从课程章节列表中删除
	err := cls.classDB.ZRem(ctx, BuildClassChapterList(*chapter.SourceId), chapter.Chid).Err()
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

	list := make([]*Chapter, len(chids)-1)

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

func (cls *ClassServer) UpdateResource(ctx context.Context, resource *Resource) error {
	return cls.chapterServer.UpdateResource(ctx, resource)
}

func (cls *ClassServer) DeleteResource(ctx context.Context, fid string, chid string) error {
	return cls.chapterServer.DeleteResource(ctx, fid, chid)
}

func (cls *ClassServer) updateClassInfo(oldClass, newClass *Class) *Class {
	if newClass.ClassName != nil {
		oldClass.WithClassName(*newClass.ClassName)
	}
	if newClass.ClassDesc != nil {
		oldClass.WithClassDesc(*newClass.ClassDesc)
	}

	if newClass.ClassType != nil {
		oldClass.WithClassType(*newClass.ClassType)
	}

	if newClass.Archive != nil {
		oldClass.WithArchive(*newClass.Archive)
	}

	if newClass.Deleted != nil {
		oldClass.WithDeleted(*newClass.Deleted)
	}

	if newClass.ClassScore != nil {
		oldClass.WithClassScore(*newClass.ClassScore)
	}

	if newClass.ClassTime != nil {
		oldClass.WithClassTime(*newClass.ClassTime)
	}

	if newClass.ClassCollege != nil {
		oldClass.WithClassCollege(*newClass.ClassCollege)
	}

	if newClass.ClassSchoolTerm != nil {
		oldClass.WithClassSchoolTerm(*newClass.ClassSchoolTerm)
	}

	if newClass.ClassOutline != nil {
		oldClass.WithClassOutline(*newClass.ClassOutline)
	}

	return oldClass
}

func (cls *ClassServer) QueryClassList(ctx context.Context, uid string) ([]*Class, error) {
	// 查询用户的课程列表
	cids, err := cls.classDB.ZRange(ctx, BuildTeacherClassList(uid), 0, -1).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|QueryClassList|QueryClassListError|%v", err)
		return nil, err
	}

	list := make([]*Class, len(cids)-1)
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

	list := make([]*Class, len(cids)-1)

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

}
