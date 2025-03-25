package http

import (
	"context"
	"cos/api/config"
	server2 "cos/api/http/internal/server"
	"errors"
	"fmt"
	"github.com/labstack/echo"
)

func StartCosHttpServer(ctx context.Context) error {
	//var configPath = flag.String("c", "E:\\code\\Go\\online_study_server01\\cos\\api\\config.json\\config.json.json", "config.json file path")

	cosServer, err := server2.NewCosServer()
	if err != nil {
		panic(err)
	}
	e := echo.New()

	server2.RegisterRouter(e, cosServer)

	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))

	return errors.New("http service stop")
}
