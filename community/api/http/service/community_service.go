package communityService

import (
	"community/api/core"
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
)

var lg = logx.GetLogger("study")

type CommunityService struct {
	ctx         context.Context
	plateServ   *core.PlateServer
	articleServ *core.ArticleServer
}

func NewCommunityService(ctx context.Context, plate *core.PlateServer, article *core.ArticleServer) *CommunityService {
	return &CommunityService{
		ctx:         ctx,
		plateServ:   plate,
		articleServ: article,
	}
}

func (cs *CommunityService) HandleCreatePlate(ctx echo.Context) error {
	decoder := json.NewDecoder(ctx.Request().Body)
	var plate core.Plate
	if err := decoder.Decode(&plate); err != nil {
		lg.Errorf("HandleCreatePlate|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.plateServ.CreatePlate(ctx.Request().Context(), &plate)
	if err != nil {
		lg.Errorf("HandleCreatePlate|CreatePlate err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreatePlate Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, plate)
}

func (cs *CommunityService) HandleUpdatePlate(ctx echo.Context) error {
	decoder := json.NewDecoder(ctx.Request().Body)
	var plate core.Plate
	if err := decoder.Decode(&plate); err != nil {
		lg.Errorf("HandleUpdatePlate|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.plateServ.UpdatePlate(ctx.Request().Context(), &plate)
	if err != nil {
		lg.Errorf("HandleUpdatePlate|UpdatePlate err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdatePlate Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, plate)
}

func (cs *CommunityService) HandleDeletePlate(ctx echo.Context) error {
	// todo 比较复杂
	return nil
}
func (cs *CommunityService) HandleListPlate(ctx echo.Context) error {
	list, err := cs.plateServ.ListPlate(ctx.Request().Context())
	if err != nil {
		lg.Errorf("HandleListPlate|ListPlate err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "ListPlate Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cs *CommunityService) HandlePublishArticle(ctx echo.Context) error {
	var article core.Article
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&article); err != nil {
		lg.Errorf("HandlePublishArticle|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	if err := cs.articleServ.CreateArticle(ctx.Request().Context(), &article); err != nil {
		lg.Errorf("HandlePublishArticle|CreateArticle err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreateArticle Error",
		})
	}
}
