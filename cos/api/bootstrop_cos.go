package main

import (
	"cos/api/config"
	"cos/api/internal/server"
	"flag"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
)

func main() {
	var configPath = flag.String("c", "./config/config.json", "config file path")

	err := config.RefreshEtcdConfig(*configPath)
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("main|RefreshEtcdConfig|err:%v", err)
		return
	}

	err = config.LoadConfigFromEtcd()
	if err != nil {
		panic(err)
	}

	cosServer, err := server.NewCosServer()
	if err != nil {
		panic(err)
	}
	e := echo.New()

	server.RegisterRouter(e, cosServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
}
