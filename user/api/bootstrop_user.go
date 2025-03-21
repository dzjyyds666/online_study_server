package main

import (
	"flag"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"user/api/config"
	"user/api/internal/server"
)

// user 启动类
func main() {
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server\\user\\api\\config.json\\config.json.json", "config.json file path")
	var configPath = flag.String("c", "/Users/zhijundu/GolandProjects/online_study_server/user/api/config/config.json", "config.json file path")
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
	userServer, err := server.NewUserServer()
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return
	}
	server.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
}
