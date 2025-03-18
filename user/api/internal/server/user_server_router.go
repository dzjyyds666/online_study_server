package server

import (
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"strings"
	"time"
	mymiddleware "user/api/internal/middleware"
)

func RegisterRouter(e *echo.Echo, us *UserServer) {

	// 跨域
	e.Use(middleware.CORS())

	// 输出当前路由的信息
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","method":"${method}","uri":"${uri}","status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	}))

	e.Use(middleware.Recover())

	globApiPrefix := e.Group("/v1")
	globApiPrefix.Add("POST", "/user/signin", us.HandlerLogin)
	globApiPrefix.Add("POST", "/user/signup", us.HandleSignUp)

	adminGroup := globApiPrefix.Group("/user/admin")
	// token验证中间件
	adminGroup.Add("GET", "/list", us.HandlerListUsers, mymiddleware.AuthMw(UserRole.Admin))
	adminGroup.Add("GET", "/delete", us.HandlerDeleteUser)

	userGroup := globApiPrefix.Group("/user")
	userGroup.Add("POST", "/update/:uid", us.UpdateUserInfo)
	userGroup.Add("GET", "/info/:uid", us.HandlerQueryUserInfo, mymiddleware.AuthMw(UserRole.Student))

	router := FilterRouter(e.Routes())

	RecordRouteToFile(router)
}

// 过滤系统路由
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
