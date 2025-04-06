package classHttpService

import (
	core2 "class/api/core"
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
	classServ *core2.ClassServer
	ctx       context.Context
}

func NewClassServer(ctx context.Context, dsClient *redis.Client) (*ClassService, error) {
	return &ClassService{
		classServ: core2.NewClassServer(ctx, dsClient),
		ctx:       ctx,
	}, nil
}

/*
只需要传递3个参数

	    ClassName  string `json:"class_name" gorm:"class_name"`
		ClassType  string `json:"class_type" gorm:"class_type"`
		ClassDesc  string `json:"class_desc" gorm:"class_desc"`
*/
func (cls *ClassService) HandleCreateClass(ctx echo.Context) error {
	var class *core2.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&class); err != nil {
		logx.GetLogger("study").Errorf("HandleCreateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	tuid := ctx.Get("uid")
	//生成随机的classId
	cid := core2.NewClassId(8)
	class.WithCid(cid).
		WithCreateTs(time.Now().Unix()).
		WithDeleted(false).
		WithArchive(false).
		WithTeacher(tuid.(string)).WithStudyClass(*class.ClassName + "教学班")

	err := cls.classServ.CreateClass(ctx.Request().Context(), class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCreateClass|CreateClass Error|%v", err)
		return err
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleCopyClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleCopyClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	info, err := cls.classServ.CopyClass(ctx.Request().Context(), cid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleCopyClass|CopyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "copy Class Error",
		})
	}
	logx.GetLogger("study").Infof("HandleCopyClass|CopyClass Success|%s", common.ToStringWithoutError(info))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleListTeacherClass(ctx echo.Context) error {
	uid := ctx.Get("uid")
	list, err := cls.classServ.QueryClassList(ctx.Request().Context(), uid.(string))
	if err != nil {
		logx.GetLogger("study").Errorf("HandleListTeacherClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}

	logx.GetLogger("study").Infof("HandleListTeacherClass|QueryClassList Success|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}
func (cls *ClassService) HandleUpdateClass(ctx echo.Context) error {
	var class *core2.Class
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

	err = cls.classServ.UpdateClass(ctx.Request().Context(), class)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUpdateClass|UpdateClass Error|%v", err)
		return err
	}
	logx.GetLogger("study").Infof("HandleUpdateClass|UpdateClass Success|%s", common.ToStringWithoutError(class))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleQueryTeacherDeletedClassList(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)

	list, err := cls.classServ.QueryTeacherDeletedClassList(ctx.Request().Context(), uid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleListTeacherClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}

	logx.GetLogger("study").Infof("HandleListTeacherClass|QueryClassList Success|%s", common.ToStringWithoutError(list))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandlePutClassInTrash(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandlePutClassInTrash|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	err := cls.classServ.MoveClassToTrash(ctx.Request().Context(), cid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandlePutClassInTrash|PutClassInTrash Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Put Class In Trash Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Put Class In Trash Success",
		"cid": cid,
	})
}

func (cls *ClassService) HandleRecoverClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleRecoverClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	err := cls.classServ.RecoverClass(ctx.Request().Context(), cid)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleRecoverClass|RecoverClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Recover Class Error",
		})
	}

	logx.GetLogger("study").Infof("HandleRecoverClass|RecoverClass Success|%s", cid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Recover Class Success",
		"cid": cid,
	})
}

func (cls *ClassService) HandleDeleteClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	err := cls.classServ.DeleteClassFromTrash(ctx.Request().Context(), cid)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleDeleteClass|Delete Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class Error",
		})
	}

	logx.GetLogger("study").Infof("HandleDeleteClass|Delete Class Success|%s", cid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleQueryClassInfo(ctx echo.Context) error {
	cid := ctx.Param("cid")

	info, err := cls.classServ.QueryClassInfo(ctx.Request().Context(), cid)
	if nil != err {
		logx.GetLogger("study").Errorf("HandleQueryClassInfo|Query Class Info Error|%v", err)
		return err
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleCreateChapter(ctx echo.Context) error {
	var chapter *core2.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil {
		logx.GetLogger("study").Errorf("HandleListUsers|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 生成章节id
	chid := core2.NewChapterId(8)
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
	var chapter *core2.Chapter
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
	var chapter *core2.Chapter
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
	var resource *core2.Resource
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
	var resource *core2.Resource
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
	var resource *core2.Resource
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

func (cls *ClassService) HandleImportStudentFromExcel(ctx echo.Context) error {
	cid := ctx.Param("cid")
	file, err := ctx.FormFile("file")
	if err != nil || len(cid) <= 0 {
		logx.GetLogger("study").Errorf("HandleImportStudentFromExcel|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	filename := file.Filename
	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleImportStudentFromExcel|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Open File Error",
		})
	}

	list, err := cls.classServ.ImportStudentFromExcel(ctx.Request().Context(), filename, cid, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleImportStudentFromExcel|ImportStudentFromExcel err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "ImportStudentFromExcel Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleUploadClassCover(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadClassCover|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	md5 := ctx.QueryParam("md5")
	fileType := ctx.QueryParam("file_type")
	dirId := ctx.QueryParam("dir_id")

	open, err := file.Open()
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadClassCover|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Open File Error",
		})
	}

	fid, err := cls.classServ.UploadClassCover(ctx.Request().Context(), md5, fileType, dirId, open)
	if err != nil {
		logx.GetLogger("study").Errorf("HandleUploadClassCover|UploadClassCover err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "UploadClassCover Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"fid": fid,
	})
}
