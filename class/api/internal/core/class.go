package core

import (
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
)

type Class struct {
	Cid          *string      `json:"cid"`
	ClassName    *string      `json:"class_name"`
	ClassDesc    *string      `json:"class_desc"`
	ClassType    *string      `json:"class_type"`
	CreateTs     *int64       `json:"create_ts"`
	Teacher      *string      `json:"teacher"`
	Archive      *bool        `json:"archive"`
	Deleted      *bool        `json:"deleted"`
	ChapterLists []Chapter    `json:"chapter_lists"` // 章节
	StudyClass   []StudyClass `json:"study_class"`   // 教学班
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

	err = json.Unmarshal([]byte(result), cl)
	if err != nil {
		logx.GetLogger("study").Errorf("QueryClassInfo|Unmarshal Class Info Error|%v", err)
		return nil, err
	}

	return cl, nil
}
