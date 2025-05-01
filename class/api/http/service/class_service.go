package classHttpService

import (
	core2 "class/api/core"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

var lg = logx.GetLogger("study")

type ClassService struct {
	classServ *core2.ClassServer
	ctx       context.Context
}

func NewClassService(ctx context.Context, dsClient *redis.Client, client *mongo.Client) (*ClassService, error) {
	return &ClassService{
		classServ: core2.NewClassServer(ctx, dsClient, client),
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
		lg.Errorf("HandleCreateClass|ctx.Bind err:%v", err)
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
		lg.Errorf("HandleCreateClass|CreateClass Error|%v", err)
		return err
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleCopyClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		lg.Errorf("HandleCopyClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	info, err := cls.classServ.CopyClass(ctx.Request().Context(), cid)
	if err != nil {
		lg.Errorf("HandleCopyClass|CopyClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "copy Class Error",
		})
	}
	lg.Infof("HandleCopyClass|CopyClass Success|%s", common.ToStringWithoutError(info))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleListTeacherClass(ctx echo.Context) error {
	uid := ctx.Get("uid")
	list, err := cls.classServ.QueryClassList(ctx.Request().Context(), uid.(string))
	if err != nil {
		lg.Errorf("HandleListTeacherClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}

	lg.Infof("HandleListTeacherClass|QueryClassList Success|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}
func (cls *ClassService) HandleUpdateClass(ctx echo.Context) error {
	var class core2.Class
	decoder := json.NewDecoder(ctx.Request().Body)
	err := decoder.Decode(&class)
	if err != nil {
		lg.Errorf("HandleUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	if class.Cid == nil {
		lg.Errorf("HandleUpdateClass|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "classid can not be null",
		})
	}

	err = cls.classServ.UpdateClass(ctx.Request().Context(), &class)
	if err != nil {
		lg.Errorf("HandleUpdateClass|UpdateClass Error|%v", err)
		return err
	}
	lg.Infof("HandleUpdateClass|UpdateClass Success|%s", common.ToStringWithoutError(class))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, class)
}

func (cls *ClassService) HandleQueryTeacherDeletedClassList(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)

	list, err := cls.classServ.QueryTeacherDeletedClassList(ctx.Request().Context(), uid)
	if err != nil {
		lg.Errorf("HandleListTeacherClass|QueryClassList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query Class List Error",
		})
	}

	lg.Infof("HandleListTeacherClass|QueryClassList Success|%s", common.ToStringWithoutError(list))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandlePutClassInTrash(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		lg.Errorf("HandlePutClassInTrash|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	err := cls.classServ.MoveClassToTrash(ctx.Request().Context(), cid)
	if err != nil {
		lg.Errorf("HandlePutClassInTrash|PutClassInTrash Error|%v", err)
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
		lg.Errorf("HandleRecoverClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	err := cls.classServ.RecoverClass(ctx.Request().Context(), cid)
	if err != nil {
		lg.Errorf("HandleRecoverClass|RecoverClass Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Recover Class Error",
		})
	}

	lg.Infof("HandleRecoverClass|RecoverClass Success|%s", cid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "Recover Class Success",
		"cid": cid,
	})
}

func (cls *ClassService) HandleDeleteClass(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		lg.Errorf("HandleDeleteClass|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	err := cls.classServ.DeleteClassFromTrash(ctx.Request().Context(), cid)
	if nil != err {
		lg.Errorf("HandleDeleteClass|Delete Class Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Class Error",
		})
	}

	lg.Infof("HandleDeleteClass|Delete Class Success|%s", cid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleQueryClassInfo(ctx echo.Context) error {
	cid := ctx.Param("cid")

	info, err := cls.classServ.QueryClassInfo(ctx.Request().Context(), cid)
	if nil != err {
		lg.Errorf("HandleQueryClassInfo|Query Class Info Error|%v", err)
		return err
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleCreateChapter(ctx echo.Context) error {
	var chapter *core2.Chapter
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&chapter); err != nil {
		lg.Errorf("HandleListUsers|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	// 生成章节id
	chid := core2.NewChapterId(8)
	chapter.Chid = &chid

	err := cls.classServ.CreateChapter(ctx.Request().Context(), chapter)
	if err != nil {
		lg.Errorf("HandleCreateChapter|Create Chapter Error|%v", err)
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
		lg.Errorf("HandleRenameChapter|ctx.Bind err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.UpdateChapter(ctx.Request().Context(), chapter)
	if err != nil {
		lg.Errorf("HandleRenameChapter|Rename Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Rename Chapter Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleDeleteChapter(ctx echo.Context) error {
	chid := ctx.Param("chid")
	err := cls.classServ.DeleteChapter(ctx.Request().Context(), chid)
	if nil != err {
		lg.Errorf("HandleDeleteChapter|Delete Chapter Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Delete Chapter Error",
		})
	}
	lg.Infof("HandleDeleteChapter|Delete Chapter Success|%s", chid)
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, nil)
}

func (cls *ClassService) HandleQueryClassChapterlist(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		lg.Errorf("HandlerQueryClassChapterlist|cid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "cid is empty",
		})
	}

	role := ctx.Get("role").(int)
	list, err := cls.classServ.QueryChapterList(ctx.Request().Context(), cid, role)
	if nil != err {
		lg.Errorf("HandlerQueryClassChapterlist|QueryChapterList|Err|%v", err)
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
		lg.Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	lg.Errorf("HandleUploadResource|resource:%s", common.ToStringWithoutError(resource))
	resource.WithPublished(false).WithDownloadable(false)
	err := cls.classServ.CreateResource(ctx.Request().Context(), resource)
	if err != nil {
		lg.Errorf("HandleUploadResource|CreateUploadResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreateUploadResource Error",
		})
	}
	lg.Infof("HandleUploadResource|CreateUploadResource|Succ|%s", common.ToStringWithoutError(resource))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, resource)
}

func (cls *ClassService) HandleUpdateResource(ctx echo.Context) error {
	var resource *core2.Resource
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&resource); err != nil {
		lg.Errorf("HandleUploadResource|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	info, err := cls.classServ.UpdateResource(ctx.Request().Context(), resource)
	if err != nil {
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdatePublishResource Error",
		})
	}
	lg.Infof("HandleUpdateResource|UpdateResource|Succ|%s", common.ToStringWithoutError(info))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleDeleteResource(ctx echo.Context) error {
	fid := ctx.Param("fid")
	info, err := cls.classServ.DeleteResource(ctx.Request().Context(), fid)
	if err != nil {
		lg.Errorf("HandleDeleteResource|DeleteResource Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteResource Error",
		})
	}
	lg.Infof("HandleDeleteResource|DeleteResource|Succ|%s", common.ToStringWithoutError(info))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleListReource(ctx echo.Context) error {
	var list *core2.ResourceList
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&list); nil != err {
		lg.Errorf("HandleQueryResourceInfo|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	list.Limit = 1000
	err := cls.classServ.QueryResourceList(ctx.Request().Context(), list)
	if err != nil {
		lg.Errorf("HandleQueryResourceInfo|Query ResourceList Error|%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "Query ResourceList Error",
		})
	}
	lg.Infof("HandleQueryResourceInfo|Query ResourceList Success|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleImportStudentFromExcel(ctx echo.Context) error {
	cid := ctx.FormValue("cid")
	file, err := ctx.FormFile("file")
	if err != nil || len(cid) <= 0 {
		lg.Errorf("HandleImportStudentFromExcel|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	filename := file.Filename
	open, err := file.Open()
	if err != nil {
		lg.Errorf("HandleImportStudentFromExcel|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Open File Error",
		})
	}

	list, err := cls.classServ.ImportStudentFromExcel(ctx.Request().Context(), filename, cid, open)
	if err != nil {
		lg.Errorf("HandleImportStudentFromExcel|ImportStudentFromExcel err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "ImportStudentFromExcel Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleUploadClassCover(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		lg.Errorf("HandleUploadClassCover|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	md5 := ctx.QueryParam("md5")
	fileType := ctx.QueryParam("file_type")
	dirId := ctx.QueryParam("dir_id")
	open, err := file.Open()
	if err != nil {
		lg.Errorf("HandleUploadClassCover|Open err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Open File Error",
		})
	}
	fid, err := cls.classServ.UploadClassCover(ctx.Request().Context(), md5, fileType, dirId, open)
	if err != nil {
		lg.Errorf("HandleUploadClassCover|UploadClassCover err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "UploadClassCover Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"fid": fid,
	})
}

func (cls *ClassService) HandleCreateTask(ctx echo.Context) error {
	lg.Infof("HandleCreateTask|Start|%s", common.ToStringWithoutError(ctx.Request().Body))
	var task core2.Task
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&task); err != nil {
		lg.Errorf("HandleCreateTask|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.CreateTask(ctx.Request().Context(), &task)
	if err != nil {
		lg.Errorf("HandleCreateTask|CreateTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "CreateTask Error",
		})
	}

	lg.Infof("HandleCreateTask|CreateTask|Succ|%s", common.ToStringWithoutError(task))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, task)
}

func (cls *ClassService) HandleListTask(ctx echo.Context) error {
	var list core2.ListTask
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&list); err != nil {
		lg.Errorf("HandleListTask|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cls.classServ.ListTask(ctx.Request().Context(), &list)
	if err != nil {
		lg.Errorf("HandleListTask|ListTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "ListTask Error",
		})
	}
	lg.Infof("HandleListTask|ListTask|Succ|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleListOwnerTask(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)
	if len(uid) <= 0 {
		lg.Errorf("HandleListOwnerTask|uid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	lg.Infof("HandleListOwnerTask|Start|%s", uid)

	list, err := cls.classServ.ListOwnerTask(ctx.Request().Context(), uid)
	if err != nil {
		lg.Errorf("HandleListOwnerTask|ListOwnerTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "ListOwnerTask Error",
		})
	}

	lg.Infof("HandleListOwnerTask|ListOwnerTask|Succ|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleDeleteTask(ctx echo.Context) error {
	tid := ctx.Param("tid")
	if len(tid) <= 0 {
		lg.Errorf("DeleteTask|tid is empty")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}

	task, err := cls.classServ.DeleteTask(ctx.Request().Context(), tid)
	if err != nil {
		lg.Errorf("DeleteTask|DeleteTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "DeleteTask Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, task)
}

func (cls *ClassService) HandleQueryStudentList(ctx echo.Context) error {
	cid := ctx.Param("cid")
	if len(cid) <= 0 {
		lg.Errorf("QueryStudentList|Param Invalid")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	list, err := cls.classServ.QueryStudentList(ctx.Request().Context(), cid)
	if err != nil {
		lg.Errorf("QueryStudentList|QueryStudentList err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryStudentList err",
		})
	}
	lg.Infof("QueryStudentList|QueryStudentList ok|%s", common.ToStringWithoutError(list))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

type addStudent struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
	Cid  string `json:"cid"`
}

func (cls *ClassService) HandleAddStudentToClass(ctx echo.Context) error {
	var student addStudent
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&student); err != nil {
		lg.Errorf("HandleAddStudentToClass|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	lg.Infof("HandleAddStudentToClass|Params bind Success|%s", common.ToStringWithoutError(student))
	if len(student.Name) <= 0 {
		student.Name = fmt.Sprintf("学生_%s", student.Uid[len(student.Uid)-4:])
	}

	err := cls.classServ.AddStudent(ctx.Request().Context(), student.Cid, student.Uid, student.Name)
	if err != nil {
		lg.Errorf("HandleAddStudentToClass|AddStudent err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "AddStudent Error",
		})
	}
	lg.Infof("HandleAddStudentToClass|AddStudent|Succ|%s", common.ToStringWithoutError(student))
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "AddStudent Success",
	})
}

func (cls *ClassService) HandleUpdateTask(ctx echo.Context) error {
	var task core2.Task
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&task); err != nil {
		lg.Errorf("HandleUpdateTask|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	lg.Infof("HandleUpdateTask|Params bind Success|%s", common.ToStringWithoutError(task))
	err := cls.classServ.UpdateTask(ctx.Request().Context(), &task)
	if err != nil {
		lg.Errorf("HandleUpdateTask|UpdateTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "UpdateTask Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "UpdateTask Success",
	})
}

func (cls *ClassService) HandleListClass(ctx echo.Context) error {
	uid := ctx.Get("uid").(string)
	list, err := cls.classServ.ListStudentClass(ctx.Request().Context(), uid)
	if err != nil {
		lg.Errorf("HandleListClass|ListStudentClass err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "ListStudentClass Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleQueryTaskInfo(ctx echo.Context) error {
	tid := ctx.Param("tid")
	if len(tid) <= 0 {
		lg.Errorf("HandleQueryTaskInfo|Param Invalid")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	info, err := cls.classServ.QueryTaskInfo(ctx.Request().Context(), tid)
	if err != nil {
		lg.Errorf("HandleQueryTaskInfo|QueryTaskInfo err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryTaskInfo Error",
		})
	}
	lg.Infof("HandleQueryTaskInfo|QueryTaskInfo|Succ|%s", common.ToStringWithoutError(info))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cls *ClassService) HandleTaskSubmit(ctx echo.Context) error {
	var task core2.SubmitTask
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&task); err != nil {
		lg.Errorf("HandleTaskSubmit|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	uid := ctx.Get("uid").(string)
	lg.Infof("HandleTaskSubmit|uid:%s", uid)
	task.WithOwner(uid)

	lg.Infof("HandleTaskSubmit|Params bind Success|%s", common.ToStringWithoutError(task))
	err := cls.classServ.TaskSubmit(ctx.Request().Context(), &task)
	if err != nil {
		if errors.Is(err, core2.ErrTaskHasSubmit) {
			lg.Errorf("HandleTaskSubmit|TaskSubmit|HasSubmit|%v", err)
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError+1, echo.Map{
				"msg": "HasSubmit",
			})
		}
		lg.Errorf("HandleTaskSubmit|TaskSubmit err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "TaskSubmit Error",
		})
	}

	lg.Infof("HandleTaskSubmit|TaskSubmit|Succ")
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "TaskSubmit Success",
	})
}

func (cls *ClassService) HandleListStudentTask(ctx echo.Context) error {
	var list core2.ListStudentList
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&list); err != nil {
		lg.Errorf("HandleListStudentTask|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	lg.Infof("HandleListStudentTask|Params bind Success|%s", common.ToStringWithoutError(list))
	err := cls.classServ.ListStudentTask(ctx.Request().Context(), &list)
	if err != nil {
		lg.Errorf("HandleListStudentTask|ListStudentTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "ListStudentTask Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cls *ClassService) HandleGetTaskStudentNumber(ctx echo.Context) error {
	tid := ctx.Param("tid")
	if len(tid) <= 0 {
		lg.Errorf("HandleGetTaskStudentNumber|Param Invalid")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Param Invalid",
		})
	}
	number, err := cls.classServ.GetTaskListNumber(ctx.Request().Context(), tid)
	if err != nil {
		lg.Errorf("HandleGetTaskStudentNumber|GetTaskListNumber err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "GetTaskListNumber Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"number": number,
	})
}

func (cls *ClassService) HandleUpdateStudentTask(ctx echo.Context) error {
	var task core2.SubmitTask
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&task); err != nil {
		lg.Errorf("HandleUpdateStudentTask|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	lg.Infof("HandleUpdateStudentTask|Params bind Success|%s", common.ToStringWithoutError(task))

	err := cls.classServ.UpdateStudentTask(ctx.Request().Context(), &task)
	if err != nil {
		lg.Errorf("HandleUpdateStudentTask|UpdateStudentTask err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdateStudentTask Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "UpdateStudentTask Success",
	})
}
