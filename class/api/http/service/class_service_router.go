package classHttpService

import (
	mymiddleware "class/api/middleware"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
)

func RegisterRouter(e *echo.Echo, cls *ClassService) {
	//e.Use(middleware.Recover())
	globApiPrefix := e.Group("/v1/class")
	teacher := globApiPrefix.Group("/tch")
	teacher.Add("GET", "/list", cls.HandleListTeacherClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/list/deleted", cls.HandleQueryTeacherDeletedClassList, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/create", cls.HandleCreateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/copy/:cid", cls.HandleCopyClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/update", cls.HandleUpdateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/delete/ctransh/:cid", cls.HandlePutClassInTrash, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/delete/:cid", cls.HandleDeleteClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/recover/:cid", cls.HandleRecoverClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/chapter/create", cls.HandleCreateChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/chapter/rename", cls.HandleRenameChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/chapter/delete/:chid", cls.HandleDeleteChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/chapter/list/:cid", cls.HandleQueryClassChapterlist, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	teacher.Add("POST", "/resource/upload", cls.HandleUploadResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/resource/delete/:fid", cls.HandleDeleteResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/resource/update", cls.HandleUpdateResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/create", cls.HandleCreateTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/stutask/list", cls.HandleListStudentTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/task/delete/:tid", cls.HandleDeleteTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/task/student/number/:tid", cls.HandleGetTaskStudentNumber, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/student/task/update", cls.HandleUpdateStudentTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/update", cls.HandleUpdateTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/student/import", cls.HandleImportStudentFromExcel, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/student/list/:cid", cls.HandleQueryStudentList, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/student/add", cls.HandleAddStudentToClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))

	student := globApiPrefix.Group("/stu")
	student.Add("POST", "/task/list", cls.HandleListTask, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("GET", "/task/:tid", cls.HandleQueryTaskInfo, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("GET", "/task/owner/list", cls.HandleListOwnerTask, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("POST", "/task/submit", cls.HandleTaskSubmit, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("GET", "/list", cls.HandleListClass, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("GET", "/query/:cid", cls.HandleQueryClassInfo, mymiddleware.AuthMw(mymiddleware.UserRole.Student))
	student.Add("GET", "/resource/list/:chid", cls.HandleListReource, mymiddleware.AuthMw(mymiddleware.UserRole.Student))

	RecordRouteToFile(FilterRouter(e.Routes()))
}

func FilterRouter(r []*echo.Route) []*echo.Route {
	var routers []*echo.Route

	for _, route := range r {
		if !strings.Contains(route.Name, "github.com/labstack/echo") {
			// 排除框架自带的路由
			route.Name = strings.TrimSuffix(route.Name, "-fm")
			routers = append(routers, route)
		}
	}
	return routers
}

type RecordRoute struct {
	Version string        `json:"version"`
	Time    int64         `json:"time"`
	Routes  []*echo.Route `json:"routes"`
}

func RecordRouteToFile(routes []*echo.Route) {
	recordRoute := RecordRoute{
		Version: "v1",
		Time:    time.Now().Unix(),
		Routes:  routes}

	data, err := json.Marshal(recordRoute)
	if err != nil {
		logx.GetLogger("study").Errorf("RecordRouteToFile|JSON Marshal Error|%v", err)
	}
	err = os.WriteFile("router.json", data, 0644)
	if err != nil {
		logx.GetLogger("study").Errorf("RecordRouteToFile|WriteFile Error|%v", err)
	}
}
