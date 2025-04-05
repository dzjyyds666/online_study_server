package http

import (
	"context"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"user/api/config"
	"user/api/http/service"
)

// user 启动类
func StartUserHttpServer(ctx context.Context, redis *redis.Client, mysql *gorm.DB) error {
	e := echo.New()
	userServer, err := userHttpService.NewUserService(ctx, redis, mysql, nil)
	if err != nil {
		logx.GetLogger("study").Errorf("UserService|StartError|NewUserServer|err:%v", err)
		return err
	}
	userHttpService.RegisterRouter(e, userServer)
	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
	return err
}
