package main

import (
	"flag"
	"github.com/dzjyyds666/online_study_server/user/api/internal/config"
	"github.com/dzjyyds666/online_study_server/user/api/internal/handler"
	"github.com/dzjyyds666/online_study_server/user/api/internal/middleware"
	"github.com/dzjyyds666/online_study_server/user/api/internal/svc"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 使用自定义中间件
	server.Use(middleware.FormatHttpResponse)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	logx.GetLogger("user").Infof("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
