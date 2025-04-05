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
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ClassServer struct {
	redis *redis.Client
	mongo *mongo.Database
}

func NewClassServer(ctx context.Context, dsClient *redis.Client) (*ClassServer, error) {

	return &ClassServer{
		redis: dsClient,
	}, nil
}

/*
只需要传递3个参数

	    ClassName  string `json:"class_name" gorm:"class_name"`
		ClassType  string `json:"class_type" gorm:"class_type"`
		ClassDesc  string `json:"class_desc" gorm:"class_desc"`
*/
func (cls *ClassServer) HandleCreateClass(ctx echo.Context) error {
	var class core.Class
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

	err := class.CreateClass(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateClass|CreateClass Error|%v", err)
		return err
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassServer) HandleCopyClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleListClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleUpdateClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleQueryDeletedClassList(ctx echo.Context) error {
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
func (cls *ClassServer) HandlePutClassInTrash(ctx echo.Context) error {
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
func (cls *ClassServer) HandleRecoverClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleDeleteClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleQueryClassInfo(ctx echo.Context) error {
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

func (cls *ClassServer) HandleCreateChapter(ctx echo.Context) error {
	var chapter core.Chapter
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

	err := chapter.CreateChapter(ctx.Request().Context(), *chapter.SourceId, cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateChapter|Create Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Create Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, chapter)
}

func (cls *ClassServer) HandleRenameChapter(ctx echo.Context) error {
	var chapter core.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil {
		logx.GetLogger("study").Errorf("HandleRenameChapter|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	info, err := chapter.RanameChapter(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRenameChapter|Rename Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Rename Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassServer) HandleDeleteChapter(ctx echo.Context) error {
	chid := ctx.Param("chid")
	if len(chid) <= 0 {
		logx.GetLogger("study").Errorf("HandleDeleteChapter|chid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	chapter := core.Chapter{Chid: &chid}
	err := chapter.DeleteChapter(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteChapter|Delete Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Chapter Error",
		})
	}
	logx.GetLogger("study").Infof("HandleDeleteChapter|Delete Chapter Success|%s", common.ToStringWithoutError(chapter))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, chapter)
}

func (cls *ClassServer) HandleQueryClassChapterlist(ctx echo.Context) error {
	decoder := json.NewDecoder(ctx.Request().Body)
	var chapterList core.ChapterList
	if err := decoder.Decode(&chapterList); err != nil {
		logx.GetLogger("study").Errorf("HandlerQueryClassChapterlist|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	list, err := chapterList.QueryChapterList(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandlerQueryClassChapterlist|QueryChapterList|Err|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryChapterList Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassServer) HandleUploadResource(ctx echo.Context) error {

	var resource core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	resource.WithPublished(false).WithDownloadable(false)

	err := resource.CreateUploadResource(ctx.Request().Context(), cls.redis)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|CreateUploadResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreateUploadResource Error",
		})
	}

	logx.GetLogger("study").Infof("HandleUploadResource|CreateUploadResource|Succ|%s", common.ToStringWithoutError(resource))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassServer) HandleUpdatePublish(ctx echo.Context) error {
	var resource core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); err != nil {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	info, err := resource.UpdatePublishResource(ctx.Request().Context(), cls.redis)
	if err != nil {
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdatePublishResource Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassServer) HandleUpdateDownloadable(ctx echo.Context) error {
	var resource core.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); nil != err {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	info, err := resource.UpdateDownloadableResource(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "update downloadable status error",
		})
	}

	logx.GetLogger("study").Infof("HandleUploadResource|UpdateDownloadableResource|Succ|%s", common.ToStringWithoutError(resource))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassServer) HandleDeleteResource(ctx echo.Context) error {
	fid := ctx.Param("fid")
	if len(fid) < 0 {
		logx.GetLogger("study").Errorf("HandleDeleteResource|fid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	resource := core.Resource{Fid: fid}
	info, err := resource.QueryResourceInfo(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Infof("HandleDeleteResource|QueryResourceInfo Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryResourceInfo Error",
		})
	}

	// 删除章节资源列表中的fid
	err = info.DeleteFormChapterList(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteResource|DeleteFormChapterList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteFormChapterList Error",
		})
	}

	// 删除资源的信息
	err = info.DeleteResource(ctx.Request().Context(), cls.redis)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteResource|DeleteResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteResource Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassServer) HandleCreateStudyClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleQueryStudyClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleDeleteStudyClass(ctx echo.Context) error {
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

func (cls *ClassServer) HandleQueryResourceInfo(ctx echo.Context) error {
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

func (cls *ClassServer) HandleImportStudyClass(ctx echo.Context) error {

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
