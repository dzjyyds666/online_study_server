package service

import (
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
	"strings"
	"time"
)

func RegisterRouter(e *echo.Echo, cs *CosService) {
	e.Use(middleware.Recover())
	cos := e.Group("/v1/cos")
	cos.Add("POST", "/upload/apply", cs.HandleApplyUpload)
	cos.Add("POST", "/upload/single/:fid", cs.HandleSingleUpload)
	cos.Add("POST", "/upload/init", cs.HandleInitMultipartUpload)
	cos.Add("POST", "/upload/part", cs.HandleUploadPart)
	cos.Add("POST", "/upload/init/video", cs.HandleInitUploadVideo)
	cos.Add("POST", "/upload/part/video", cs.HandleUploadVideoPart)
	cos.Add("POST", "/upload/complete/:fid", cs.CompleteUpload)
	cos.Add("POST", "/upload/abort/:fid", cs.HandleAbortUpload)

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
