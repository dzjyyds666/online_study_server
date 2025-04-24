package communityService

import (
	"community/api/core"
	"community/api/middleware"
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
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, article)
}

func (cs *CommunityService) HandleUpdateArticle(ctx echo.Context) error {
	decoder := json.NewDecoder(ctx.Request().Body)
	var article core.Article
	if err := decoder.Decode(&article); err != nil {
		lg.Errorf("HandleUpdateArticle|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	if len(article.Status) > 0 {
		// 判断当前的用户是不是管理员
		role := ctx.Get("role").(int)
		if role != middleware.UserRole.Admin {
			return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpForbidden, echo.Map{
				"msg": "No Permission",
			})
		}
	}
	if err := cs.articleServ.UpdateArticle(ctx.Request().Context(), &article); err != nil {
		lg.Errorf("HandleUpdateArticle|UpdateArticle err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "UpdateArticle Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, article)
}

func (cs *CommunityService) HandleDeleteArticle(ctx echo.Context) error {
	aid := ctx.Param("aid")
	if len(aid) <= 0 {
		lg.Errorf("HandleDeleteArticle|Param Invalid")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}
	if err := cs.articleServ.DeleteArticle(ctx.Request().Context(), aid); err != nil {
		lg.Errorf("HandleDeleteArticle|DeleteArticle err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteArticle Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "DeleteArticle Succ",
	})
}

func (cs *CommunityService) HandleListArticle(ctx echo.Context) error {
	var list core.ListArticle
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&list); err != nil {
		lg.Errorf("HandleListArticle|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	if err := cs.articleServ.ListArticle(ctx.Request().Context(), &list, ctx.Get("role").(int)); err != nil {
		lg.Errorf("HandleListArticle|ListArticle err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "ListArticle Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}
