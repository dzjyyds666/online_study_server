package http

import (
	"context"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"user/api/config"
	"user/api/core"
	"user/api/http/service"
)

// user 启动类
func StartUserHttpServer(ctx context.Context, server *core.UserServer) error {
	e := echo.New()
	userServer, err := userHttpService.NewUserService(ctx, server)
	if err != nil {
		logx.GetLogger("study").Errorf("UserService|StartError|NewUserServer|err:%v", err)
		return err
	}
	userHttpService.RegisterRouter(e, userServer)
	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
	return err
}
