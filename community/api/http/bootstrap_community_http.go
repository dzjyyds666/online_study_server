package http

import (
	"community/api/config"
	"community/api/core"
	communityService "community/api/http/service"
	"context"
	"fmt"
	"github.com/labstack/echo"
)

func StartCommunityHttpServer(ctx context.Context, plate *core.PlateServer, article *core.ArticleServer, comment *core.CommentServer) {
	e := echo.New()
	communityServer := communityService.NewCommunityService(ctx, plate, article, comment)
	communityService.RegisterRouter(e, communityServer)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", config.GloableConfig.Port)))
	return
}
