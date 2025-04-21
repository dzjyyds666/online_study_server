package http

import (
	"context"
	"cos/api/config"
	"cos/api/http/service"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

func StartCosHttpServer(ctx context.Context, ds *redis.Client, s3Client *s3.Client) error {
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server01\\cos\\common\\config.json\\config.json.json", "config.json file path")

	cosServer, err := service.NewCosServer(ctx, ds, s3Client)
	if err != nil {
		panic(err)
	}
	e := echo.New()

	service.RegisterRouter(e, cosServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))

	return errors.New("http service stop")
}
