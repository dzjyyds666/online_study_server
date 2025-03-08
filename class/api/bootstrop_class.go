package main

import (
	"class/api/config"
	"class/api/internal/server"
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
	e := echo.New()
	userServer, err := server.NewClassServer()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return
	}
	server.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
}
