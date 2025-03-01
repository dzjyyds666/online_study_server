package main

import (
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"user/config"
	"user/internal/server"
)

// user 启动类
func main() {
	err := config.LoadConfigFromEtcd()
	if err != nil {
		panic(err)
	}
	e := echo.New()
	userServer, err := server.NewUserServer()
	if err != nil {
		logx.GetLogger("OS_Server").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return
	}
	server.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
}
