package http

import (
	"community/api/config"
	communityService "community/api/http/service"
	"context"
	"fmt"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

func StartCommunityHttpServer(ctx context.Context, ds *redis.Client) {
	e := echo.New()
	communityServer := communityService.NewCommunityService(ctx, ds)
	communityService.RegisterRouter(e, communityServer)
	e.Logger.Fatal(e.Start(fmt.Sprint(":", *config.GloableConfig.Port)))
	return
}
