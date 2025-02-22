package main

import (
	"flag"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github/dzjyyds666/online_study_server/user/api/internal/config"
	"github/dzjyyds666/online_study_server/user/api/internal/handler"
	"github/dzjyyds666/online_study_server/user/api/internal/svc"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	logx.GetLogger("user").Infof("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
