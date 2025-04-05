package http

import (
	"context"
	"cos/api/config"
	"cos/api/http/service"
	"errors"
	"fmt"
	"github.com/labstack/echo"
)

func StartCosHttpServer(ctx context.Context) error {
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server01\\cos\\common\\config.json\\config.json.json", "config.json file path")

	cosServer, err := service.NewCosServer()
	if err != nil {
		panic(err)
	}
	e := echo.New()

	service.RegisterRouter(e, cosServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))

	return errors.New("http service stop")
}
