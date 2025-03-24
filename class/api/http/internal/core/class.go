package core

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
	"time"
)

type Class struct {
	Cid            *string      `json:"cid,omitempty"`
	ClassName      *string      `json:"class_name,omitempty"`
	ClassDesc      *string      `json:"class_desc,omitempty"`
	ClassType      *string      `json:"class_type,omitempty"`
	CreateTs       *int64       `json:"create_ts,omitempty"`
	Teacher        *string      `json:"teacher,omitempty"`
	Archive        *bool        `json:"archive,omitempty"`
	Deleted        *bool        `json:"deleted,omitempty"`
	StudyClassList []StudyClass `json:"study_class_list,omitempty"`
	ChapterList    []Chapter    `json:"chapter_list,omitempty"`
}

func (ci *Class) WithCid(id string) *Class {
	ci.Cid = &id
	return ci
}

func (ci *Class) WithClassName(name string) *Class {
	ci.ClassName = &name
	return ci
}

func (ci *Class) WithClassDesc(desc string) *Class {
	ci.ClassDesc = &desc
	return ci
}

func (ci *Class) WithClassType(type_ string) *Class {
	ci.ClassType = &type_
	return ci
}

func (ci *Class) WithCreateTs(ts int64) *Class {
	ci.CreateTs = &ts
	return ci
}

func (ci *Class) WithTeacher(teacher string) *Class {
	ci.Teacher = &teacher
	return ci
}

func (ci *Class) WithArchive(archive bool) *Class {
	ci.Archive = &archive
	return ci
}

func (ci *Class) WithDeleted(deleted bool) *Class {
	ci.Deleted = &deleted
	return ci
}

type UpdateChapters struct {
	Delete   bool
	Chapters []Chapter
}

type UpdateStudyClass struct {
	Delete     bool
	StudyClass []StudyClass
}

// todo 如何更加优雅的实现
func (ci *Class) UpdateClasInfo(ctx context.Context, ds *redis.Client) error {
	// 先去redis中获取原始的课程信息
	infoKey := BuildClassInfo(*ci.Cid)
	result, err := ds.Get(ctx, infoKey).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryCLassInfo|Get Class Info Error|%v", err)
		return err
	}

	var originInfo Class
	if err := json.Unmarshal([]byte(result), &originInfo); err != nil {
		logx.GetLogger("study").Errorf("QueryCLassInfo|Unmarshal Class Info Error|%v", err)
		return err
	}

	// 先替换所有的字段
	if ci.ClassName != nil {
		originInfo.ClassName = ci.ClassName
	}

	if ci.ClassDesc != nil {
		originInfo.ClassDesc = ci.ClassDesc
	}

	if ci.ClassType != nil {
		originInfo.ClassType = ci.ClassType
	}

	if ci.Archive == nil {
		originInfo.Archive = ci.Archive
	}

	if ci.Deleted == nil {
		originInfo.Deleted = ci.Deleted
	}

	// 重新写入reids
	classInfoStr, err := json.Marshal(originInfo)
	if err != nil {
		logx.GetLogger("study").Errorf("QueryCLassInfo|Marshal Class Info Error|%v", err)
		return err
	}

	err = ds.Set(ctx, infoKey, string(classInfoStr), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("QueryCLassInfo|Set Class Info Error|%v", err)
		return err
	}
	return nil
}

func (cl *Class) QueryClassInfo(ctx context.Context, ds *redis.Client) (*Class, error) {
	classInfoKey := BuildClassInfo(*cl.Cid)

	result, err := ds.Get(ctx, classInfoKey).Result()
	if nil != err {
		logx.GetLogger("study").Errorf("QueryClassInfo|Get Class Info Error|%v", err)
		return nil, err
	}

	var classInfo Class
	err = json.Unmarshal([]byte(result), &classInfo)
	if err != nil {
		logx.GetLogger("study").Errorf("QueryClassInfo|Unmarshal Class Info Error|%v", err)
		return nil, err
	}

	return &classInfo, nil
}

func (cl *Class) DeleteFromTrash(ctx context.Context, ds *redis.Client) (*Class, error) {
	// 获取到classinfo
	result, err := ds.Get(ctx, BuildClassInfo(*cl.Cid)).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(result), cl)
	if err != nil {
		return nil, err
	}

	if *cl.Deleted == false {
		return nil, errors.New("class not in trash")
	}

	// 遍历删除课程下面的章节
	chids, err := ds.ZRange(ctx, BuildSourceChapterList(*cl.Cid), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	for _, chid := range chids {
		chapter := Chapter{Chid: &chid}
		err := chapter.DeleteChapter(ctx, ds)
		if err != nil {
			logx.GetLogger("study").Errorf("DeleteFromTrash|Delete Chapter Error|%v", err)
			return nil, err
		}
	}

	if err := ds.ZRem(ctx, BuildClassDeletedList(), *cl.Cid).Err(); err != nil {
		logx.GetLogger("study").Errorf("DeleteFromTrash|Delete Class From Deleted List Error|%v", err)
		return nil, err
	}

	if err = ds.ZRem(ctx, BuildTeacherClassDeletedList(*cl.Teacher), *cl.Cid).Err(); err != nil {
		logx.GetLogger("study").Errorf("DeleteFromTrash|Delete Class From Teacher Deleted List Error|%v", err)
		return nil, err
	}

	err = ds.Del(ctx, BuildClassInfo(*cl.Cid)).Err()
	if nil != err {
		logx.GetLogger("study").Errorf("DeleteFromTrash|Delete Class Info Error|%v", err)
		return nil, err
	}

	return cl, nil
}

func (cl *Class) CopyClass(ctx context.Context, ds *redis.Client) (*Class, error) {
	newCid := NewClassId(8)

	info, err := cl.QueryClassInfo(ctx, ds)
	if err != nil {
		logx.GetLogger("study").Errorf("CopyClass|Query Class Info Error|%v", err)
		return nil, err
	}

	// 构建新的class
	info.WithCid(newCid).
		WithClassName(*cl.ClassName).
		WithArchive(false).
		WithDeleted(false).
		WithCreateTs(time.Now().Unix())
	logx.GetLogger("study").Infof("CopyClass|Copy Class|%s", common.ToStringWithoutError(info))

	// 写入创建新的的class
	err = info.CreateClass(ctx, ds)
	if err != nil {
		logx.GetLogger("study").Errorf("CopyClass|Create Class Error|%v", err)
		return nil, err
	}

	// 查询旧课程下的所有章节
	chids, err := ds.ZRange(ctx, BuildSourceChapterList(*cl.Cid), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	for _, chid := range chids {
		chapter := Chapter{Chid: &chid}
		chapterInfo, err := chapter.QueryChapterInfo(ctx, ds)
		if err != nil {
			logx.GetLogger("study").Errorf("CopyClass|Query Chapter Info Error|%v", err)
			return nil, err
		}

		// 查询章节下的所有资源
		fids, err := chapterInfo.QueryResourcList(ctx, ds, "", -1)
		if err != nil {
			logx.GetLogger("study").Errorf("CopyClass|Query Resource List Error|%v", err)
			return nil, err
		}

		newChapterId := NewChapterId(8)
		chapter.WithChid(newChapterId).WithSourceId(newCid)
		// 写入章节信息
		err = chapter.CreateChapter(ctx, newCid, ds)
		if err != nil {
			logx.GetLogger("study").Errorf("CopyClass|Create Chapter Error|%v", err)
			return nil, err
		}

		for _, fid := range fids {
			// todo 复制文件的时候，有一点问题
			resource := Resource{Fid: fid}
			resourceInfo, err := resource.QueryResourceInfo(ctx, ds)
			if err != nil {
				logx.GetLogger("study").Errorf("CopyClass|Query Resource Info Error|%v", err)
				return nil, err
			}
			resourceInfo.WithFid(NewFid()).WithChid(newChapterId)
			err = resourceInfo.CreateUploadResource(ctx, ds)
			if err != nil {
				logx.GetLogger("study").Errorf("CopyClass|Create Resource Error|%v", err)
				return nil, err
			}
		}
		info.ChapterList = append(info.ChapterList, chapter)
	}
	return info, nil
}

func (cl *Class) CreateClass(ctx context.Context, ds *redis.Client) error {
	marshal, err := json.Marshal(&cl)
	if err != nil {
		logx.GetLogger("study").Errorf("CreateClass|Marshal Class Error|%v", err)
		return err
	}

	err = ds.Set(ctx, BuildClassInfo(*cl.Cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateClass|Marshal Class Error|%v", err)
		return err
	}

	// 把该课程加入教师的课程列表中
	err = ds.ZAdd(ctx, BuildTeacherClassList(*cl.Teacher), redis.Z{
		Score:  float64(*cl.CreateTs),
		Member: *cl.Cid,
	}).Err()

	if err != nil {
		logx.GetLogger("study").Errorf("CreateClass|Add Class To Teacher List Error|%v", err)
		return err
	}

	// 把该课程加入到课程list中
	_, err = ds.ZAdd(ctx, BuildAllClassList(), redis.Z{
		Score:  float64(*cl.CreateTs),
		Member: *cl.Cid,
	}).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("CreateClass|Add Class To ClassList Error|%v", err)
		return err
	}

	return err
}
