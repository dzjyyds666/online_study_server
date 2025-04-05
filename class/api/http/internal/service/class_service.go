package classHttpService

import (
	"class/api/http/internal/core"
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"time"
)

type ClassService struct {
	redis     *redis.Client
	classServ *core.ClassServer
	ctx       context.Context
}

func NewClassServer(ctx context.Context, dsClient *redis.Client) (*ClassService, error) {
	return &ClassService{
		classServ: core.NewClassServer(ctx, dsClient),
		ctx:       ctx,
		redis:     dsClient,
	}, nil
}

/*
只需要传递3个参数

	    ClassName  string `json:"class_name" gorm:"class_name"`
		ClassType  string `json:"class_type" gorm:"class_type"`
		ClassDesc  string `json:"class_desc" gorm:"class_desc"`
*/
func (cls *ClassService) HandleCreateClass(ctx echo.Context) error {
	var class *core.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&class); err != nil {
		logx.GetLogger("study").Errorf("HandleCreateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	tuid := ctx.Get("uid")
	//生成随机的classId
	cid := core.NewClassId(8)
	class.WithCid(cid).
		WithCreateTs(time.Now().Unix()).
		WithDeleted(false).
		WithArchive(false).
		WithTeacher(tuid.(string))

	err := cls.classServ.CreateClass(ctx.Request().Context(), class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateClass|CreateClass Error|%v", err)
		return err
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleCopyClass(ctx echo.Context) error {
	var class core.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&class); err != nil {
		logx.GetLogger("study").Errorf("HandleCopyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	info, err := class.CopyClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCopyClass|CopyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "copy Class Error",
		})
	}

	logx.GetLogger("study").Infof("HandleCopyClass|CopyClass Success|%s", common.ToStringWithoutError(class))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleListClass(ctx echo.Context) error {
	var classLists core.ClassList
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&classLists); err != nil {
		logx.GetLogger("study").Errorf("HandleListClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	uid := ctx.Get("uid")
	list, err := classLists.QueryClassList(ctx.Request().Context(), uid.(string), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleListClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleUpdateClass(ctx echo.Context) error {
	var class core.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	err := decoder.Decode(&class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	if class.Cid == nil {
		logx.GetLogger("study").Errorf("HandleUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "classid can not be null",
		})
	}

	class.StudyClassList = nil
	class.ChapterList = nil

	marshal, err := json.Marshal(&class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUpdateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "system internal",
		})
	}

	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfo(*class.Cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUpdateClass|Marshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "system internal",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleQueryDeletedClassList(ctx echo.Context) error {
	decoder := json.NewDecoder(ctx.Request().Body)
	var classLists core.ClassList
	if err := decoder.Decode(&classLists); err != nil {
		logx.GetLogger("study").Errorf("HandleListClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	uid := ctx.Get("uid").(string)

	err := classLists.QueryDeletedClassList(ctx.Request().Context(), cls.redis, uid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleListClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}

	logx.GetLogger("study").Infof("HandleListClass|QueryClassList Success|%s", common.ToStringWithoutError(classLists))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, classLists)
}

// 上传完课程信息之后
// 把课程移入垃圾箱
func (cls *ClassService) HandlePutClassInTrash(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	var class core.Class
	// 把课程的删除为修改为delete
	result, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Get Class Info Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Class Info Error",
		})
	}
	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Unmarshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Unmarshal Class Error",
		})
	}

	class.WithDeleted(true)
	marshal, _ := json.Marshal(&class)
	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfo(*class.Cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	// 从classlist中移除
	err = cls.redis.ZRem(ctx.Request().Context(), core.RedisAllClassList, class.Cid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Remove Class From ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From ClassList Error",
		})
	}

	// 添加到课程删除列表
	err = cls.redis.ZAdd(ctx.Request().Context(), core.BuildClassDeletedList(), redis.Z{
		Member: class.Cid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Add Class To Teacher Deleted List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To Teacher Deleted List Error",
		})
	}

	// 从老师的正常课程列表清除
	teacherKey := core.BuildTeacherClassList(*class.Teacher)
	err = cls.redis.ZRem(ctx.Request().Context(), teacherKey, class.Cid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Remove Class From Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From Teacher List Error",
		})
	}

	// 从加入老师的删除列表中
	teacherKey = core.BuildTeacherClassDeletedList(*class.Teacher)
	err = cls.redis.ZAdd(ctx.Request().Context(), teacherKey, redis.Z{
		Member: class.Cid,
		Score:  float64(time.Now().Unix()),
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Add Class To Teacher Deleted List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To Teacher Deleted List Error",
		})
	}

	logx.GetLogger("study").Infof("HandleDeleteClass|HandleDeleteClass Success|%s", common.ToStringWithoutError(class))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

// 恢复课程
func (cls *ClassService) HandleRecoverClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleRecoverClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	var class core.Class
	result, err := cls.redis.Get(ctx.Request().Context(), core.BuildClassInfo(cid)).Result()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Get Class Info Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Get Class Info Error",
		})
	}

	err = json.Unmarshal([]byte(result), &class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Unmarshal Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Unmarshal Class Error",
		})
	}

	class.WithDeleted(false)

	marshal, _ := json.Marshal(&class)
	err = cls.redis.Set(ctx.Request().Context(), core.BuildClassInfo(cid), marshal, 0).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Update Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Update Class Error",
		})
	}

	// 把课程添加到classlist中
	err = cls.redis.ZAdd(ctx.Request().Context(), core.BuildAllClassList(), redis.Z{
		Score:  float64(*class.CreateTs),
		Member: cid,
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Add Class To ClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To ClassList Error",
		})
	}

	// 从老师的删除列表中移除，添加到正常列表中
	teacherKey := core.BuildTeacherClassDeletedList(*class.Teacher)
	err = cls.redis.ZRem(ctx.Request().Context(), teacherKey, cid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Remove Class From Teacher Deleted List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From Teacher Deleted List Error",
		})
	}

	// 从课程的删除列表中移除
	err = cls.redis.ZRem(ctx.Request().Context(), core.BuildClassDeletedList(), cid).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Remove Class From Class Deleted List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Remove Class From Class Deleted List Error",
		})
	}

	teacherKey = core.BuildTeacherClassList(*class.Teacher)
	err = cls.redis.ZAdd(ctx.Request().Context(), teacherKey, redis.Z{
		Score:  float64(*class.CreateTs),
		Member: cid,
	}).Err()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|Add Class To Teacher List Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Add Class To Teacher List Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleDeleteClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}
	uid := ctx.Get("uid").(string)

	class := core.Class{Cid: &cid, Teacher: &uid}

	info, err := class.DeleteFromTrash(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Delete Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleQueryClassInfo(ctx echo.Context) error {
	cid := ctx.Param("cid")

	class := core.Class{}
	class.WithCid(cid)

	info, err := class.QueryClassInfo(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleQueryClassInfo|Query Class Info Error|%v", err)
		return err
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleCreateChapter(ctx echo.Context) error {
	var chapter *core.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil {
		logx.GetLogger("study").Errorf("HandleListUsers|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 生成章节id
	chid := core.NewChapterId(8)
	chapter.Chid = &chid

	err := cls.classServ.CreateChapter(ctx.Request().Context(), chapter)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateChapter|Create Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, chapter)
}

func (cls *ClassService) HandleRenameChapter(ctx echo.Context) error {
	var chapter *core.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil || chapter.ChapterName == nil {
		logx.GetLogger("study").Errorf("HandleRenameChapter|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.UpdateChapter(ctx.Request().Context(), chapter)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRenameChapter|Rename Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Rename Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleDeleteChapter(ctx echo.Context) error {
	var chapter *core.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteChapter|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.DeleteChapter(ctx.Request().Context(), chapter)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteChapter|Delete Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Chapter Error",
		})
	}
	logx.GetLogger("study").Infof("HandleDeleteChapter|Delete Chapter Success|%s", *chapter.Chid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleQueryClassChapterlist(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandlerQueryClassChapterlist|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	list, err := cls.classServ.QueryChapterList(ctx.Request().Context(), cid)
	if nil != err {
		logx.GetLogger("study").Errorf("HandlerQueryClassChapterlist|QueryChapterList|Err|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryChapterList Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleUploadResource(ctx echo.Context) error {

	var resource *core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	resource.WithPublished(false).WithDownloadable(false)

	err := cls.classServ.CreateResource(ctx.Request().Context(), resource)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|CreateUploadResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreateUploadResource Error",
		})
	}

	logx.GetLogger("study").Infof("HandleUploadResource|CreateUploadResource|Succ|%s", common.ToStringWithoutError(resource))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassService) HandleUpdateResource(ctx echo.Context) error {
	var resource *core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	err := cls.classServ.UpdateResource(ctx.Request().Context(), resource)
	if err != nil {
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdatePublishResource Error",
		})
	}

	logx.GetLogger("study").Infof("HandleUpdateResource|UpdateResource|Succ|%s", common.ToStringWithoutError(resource))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleDeleteResource(ctx echo.Context) error {
	var resource *core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(resource); err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.DeleteResource(ctx.Request().Context(), *resource.Fid, *resource.Chid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteResource|DeleteResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteResource Error",
		})
	}

	logx.GetLogger("study").Infof("HandleDeleteResource|DeleteResource|Succ|%s", *resource.Fid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassService) HandleCreateStudyClass(ctx echo.Context) error {
	var studyClass core.StudyClass
	if err := ctx.Bind(&studyClass); err != nil {
		logx.GetLogger("study").Errorf("HandleCreateStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	studyClass.SCid = core.NewStudyClass(8)

	err := studyClass.CreateStudyClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateStudyClass|Create StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, studyClass)
}

func (cls *ClassService) HandleQueryStudyClass(ctx echo.Context) error {
	var studyClassList core.StudyClassList
	if err := ctx.Bind(&studyClassList); err != nil {
		logx.GetLogger("study").Errorf("HandleQueryStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	list, err := studyClassList.QueryStudyClassList(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleQueryStudyClass|Query StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleDeleteStudyClass(ctx echo.Context) error {
	var studyClass core.StudyClass
	if err := ctx.Bind(&studyClass); err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteStudyClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := studyClass.DeleteStudyClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleDeleteStudyClass|Delete StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete StudyClass Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, studyClass)
}

func (cls *ClassService) HandleQueryResourceInfo(ctx echo.Context) error {
	var resourceList core.ResourceList
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resourceList); nil != err {
		logx.GetLogger("study").Errorf("HandleQueryResourceInfo|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	list, err := resourceList.QueryResourceList(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleQueryResourceInfo|Query ResourceList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query ResourceList Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleImportStudyClass(ctx echo.Context) error {

	cid := ctx.Param("cid")
	file, err := ctx.FormFile("file")
	if err != nil || len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleImportStudyClass|FormFile err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleImportStudyClass|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	scid := core.NewStudyClass(8)
	studyClass := core.StudyClass{SCid: scid, Cid: cid}
	err = studyClass.ImportStudyClass(ctx.Request().Context(), cls.redis, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleImportStudyClass|Import StudyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Import StudyClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, studyClass)
}
