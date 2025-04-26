package communityService

import (
	"community/api/middleware"
	"encoding/json"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"os"
	"strings"
	"time"
)

func RegisterRouter(e *echo.Echo, cs *CommunityService) {
	g := e.Group("/v1/community")
	g.Add("POST", "/plate/create", cs.HandleCreatePlate, middleware.AuthMw(middleware.UserRole.Admin))
	g.Add("POST", "/plate/update", cs.HandleUpdatePlate, middleware.AuthMw(middleware.UserRole.Admin))
	g.Add("GET", "/plate/list", cs.HandleListPlate, middleware.AuthMw(middleware.UserRole.Student))
	g.Add("GET", "/plate/info/:pid", cs.HandleGetPlateInfo, middleware.AuthMw(middleware.UserRole.Student))
	g.Add("POST", "/article/create", cs.HandlePublishArticle, middleware.AuthMw(middleware.UserRole.Student))
	g.Add("POST", "/article/update", cs.HandleUpdateArticle, middleware.AuthMw(middleware.UserRole.Admin))
	g.Add("GET", "/article/delete/:aid", cs.HandleDeleteArticle, middleware.AuthMw(middleware.UserRole.Student))
	g.Add("GET", "/article/list", cs.HandleListArticle, middleware.AuthMw(middleware.UserRole.Student))

	g.Add("POST", "/comment/create", cs.HandleCreateComment, middleware.AuthMw(middleware.UserRole.Student))
	g.Add("POST", "/comment/list", cs.HandleListComment, middleware.AuthMw(middleware.UserRole.Student))

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
