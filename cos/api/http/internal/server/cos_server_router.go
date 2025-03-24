package server

import (
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"strings"
	"time"
)

func RegisterRouter(e *echo.Echo, cs *CosServer) {
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	cos := e.Group("/v1/cos")
	cos.Add("POST", "/upload/apply", cs.HandlerApplyUpload)
	cos.Add("POST", "/upload/single/:fid", cs.HandlerSingleUpload)
	cos.Add("GET", "/file/:fid", cs.HandlerGetFile)

	cos.Add("POST", "/upload/init", cs.HandlerInitMultipartUpload)
	cos.Add("POST", "/upload/multi/:fid", cs.HandlerMultiUpload)
	cos.Add("POST", "/upload/complete/:fid", cs.CompleteUpload)
	cos.Add("POST", "/upload/abort/:fid", cs.HandlerAbortUpload)

	RecordRouteToFile(FilterRouter(e.Routes()))
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
		logx.GetLogger("study").Errorf("RecordRouteToFile|JSON Marshal Error|%v", err)
	}
	err = os.WriteFile("router.json", data, 0644)
	if err != nil {
		logx.GetLogger("study").Errorf("RecordRouteToFile|WriteFile Error|%v", err)
	}
}
