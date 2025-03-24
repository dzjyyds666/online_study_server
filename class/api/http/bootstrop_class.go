package main

import (
	"class/api/config"
	"class/api/http/internal/service"
	"flag"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
)

func main() {
	var configPath = flag.String("c", "/Users/zhijundu/GolandProjects/online_study_server/class/api/config/config.json", "config.json file path")

	err := config.RefreshEtcdConfig(*configPath)
	if err != nil {
		logx.GetLogger("study").Errorf("main|RefreshEtcdConfig|err:%v", err)
		return
	}

	err = config.LoadConfigFromEtcd()
	if err != nil {
		panic(err)
	}
	e := echo.New()
	userServer, err := service.NewClassServer()
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return
	}

	service.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))

}
