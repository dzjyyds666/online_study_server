package http

import (
	"context"
	"fmt"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"user/api/config"
	"user/api/http/internal/server"
)

// user.proto 启动类
func StartUserHttpServer(ctx context.Context, redis *redis.Client, mysql *gorm.DB) error {
	e := echo.New()
	userServer, err := userHttpService.NewUserServer(ctx, redis, mysql, nil)
	if err != nil {
		logx.GetLogger("study").Errorf("UserServer|StartError|NewUserServer|err:%v", err)
		return err
	}

	userHttpService.RegisterRouter(e, userServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
	return err
}
