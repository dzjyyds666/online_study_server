package http

import (
	"class/api/config"
	"class/api/http/service"
	"context"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

func StartClassHttpServer(ctx context.Context, ds *redis.Client) error {

	e := echo.New()
	userServer, err := classHttpService.NewClassService(ctx, ds)
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return err
	}

	classHttpService.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))

	return nil

}
