package core

import (
	"class/api/http/internal/lua"
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/redis/go-redis/v9"
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

	class.WithDeleted(true)
	err = cls.classDB.Set(ctx, BuildClassInfo(cid), class.Marshal(), 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("ClassServer|MoveClassToTrash|UpdateClassInfoError|%v", err)
		return err
	}

	err = cls.classDB.Eval(ctx, lua.MoveClassToTrash, []string{
		BuildTeacherClassList(cid),
		BuildTeacherClassDeletedList(cid),
		BuildAllClassList(),
		BuildClassDeletedList(),
	}, cid, class.CreateTs).Err()

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
