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

	AuthMw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return mymiddleware.AuthVerifyMw(next, us.redis)
	}

	e.Use(middleware.Recover())

	globApiPrefix := e.Group("/v1/api")
	globApiPrefix.Add("POST", "/login", us.HandlerLogin)
	globApiPrefix.Add("POST", "/signup", us.SignUp)
	globApiPrefix.Add("Get", "/send/verifyCode", us.SendMessage)

	adminGroup := globApiPrefix.Group("/admin")
	// token验证中间件
	adminGroup.Use(AuthMw)
	adminGroup.Add("GET", "/user/list", us.HandlerListUsers)
	adminGroup.Add("GET", "/user/delete", us.HandlerDeleteUser)

	userGroup := globApiPrefix.Group("")
	userGroup.Use(AuthMw)
	userGroup.Add("GET", "/user/update", us.UpdateUserInfo)
	userGroup.Add("GET", "/user/info/:fid", us.HandlerQueryUserInfo)

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
