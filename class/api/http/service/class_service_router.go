package classHttpService

import (
	mymiddleware "class/api/middleware"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"strings"
	"time"
)

func RegisterRouter(e *echo.Echo, cls *ClassService) {
	e.Use(middleware.Recover())
	globApiPrefix := e.Group("/v1/class")
	teacher := globApiPrefix.Group("/tch")
	teacher.Add("GET", "/list", cls.HandleListTeacherClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/list/deleted", cls.HandleQueryTeacherDeletedClassList, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/create", cls.HandleCreateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/copy", cls.HandleCopyClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/update", cls.HandleUpdateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/delete/ctransh/:cid", cls.HandlePutClassInTrash, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/delete/:cid", cls.HandleDeleteClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/recover/:cid", cls.HandleRecoverClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/chapter/create", cls.HandleCreateChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/chapter/rename", cls.HandleRenameChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/chapter/delete/:chid", cls.HandleDeleteChapter, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/chapter/list/:cid", cls.HandleQueryClassChapterlist, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/resource/upload", cls.HandleUploadResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/resource/delete/:fid", cls.HandleDeleteResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/resource/update", cls.HandleUpdateResource, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/create", cls.HandleCreateTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("POST", "/task/list", cls.HandleListTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))
	teacher.Add("GET", "/task/delete/:tid", cls.HandleDeleteTask, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher))

	student := globApiPrefix.Group("/stu")
	student.Add("GET", "/query/:cid", cls.HandleQueryClassInfo, mymiddleware.AuthMw(mymiddleware.UserRole.Student))

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
