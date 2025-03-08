package server

import (
	mymiddleware "class/api/internal/middleware"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"strings"
	"time"
)

func RegisterRouter(e *echo.Echo, cls *ClassServer) {
	e.Use(middleware.Recover())

	globApiPrefix := e.Group("/v1/api")
	teacher := globApiPrefix.Group("/tch/class")
	teacher.Add("POST", "/craete", cls.HandlerCreateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("POST", "/upload", cls.HandlerClassAddVideo, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("POST", "/update", cls.HandlerUpdateClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("POST", "/delete/cvideo", cls.HandlerDeleteVideo, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("POST", "/delete/ctransh", cls.HandlerPutClassInTrash, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("Get", "/delete/:cid", cls.HandlerDeleteClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))
	teacher.Add("POST", "/recover/:cid", cls.HandlerRecoverClass, mymiddleware.AuthMw(mymiddleware.UserRole.Teacher, cls.redis))

	student := globApiPrefix.Group("/stu/class")
	student.Add("POST", "/query/:cid", cls.HandlerQueryClassInfo, mymiddleware.AuthMw(mymiddleware.UserRole.Student, cls.redis))
	student.Add("POST", "/subscribe/:cid", cls.HandlerSubscribeClass, mymiddleware.AuthMw(mymiddleware.UserRole.Student, cls.redis))
	student.Add("POST", "/cancle/:cid", cls.HandlerCancleSubscribeClass, mymiddleware.AuthMw(mymiddleware.UserRole.Student, cls.redis))

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
		logx.GetLogger("OS_Server").Errorf("RecordRouteToFile|JSON Marshal Error|%v", err)
	}
	err = os.WriteFile("router.json", data, 0644)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("RecordRouteToFile|WriteFile Error|%v", err)
	}
}
