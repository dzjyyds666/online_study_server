package communityService

import (
	"community/api/core"
	"community/api/middleware"
	"context"
	"encoding/json"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
)

var lg = logx.GetLogger("study")

type CommunityService struct {
	ctx         context.Context
	plateServ   *core.PlateServer
	articleServ *core.ArticleServer
	commentServ *core.CommentServer
}

func NewCommunityService(ctx context.Context, plate *core.PlateServer, article *core.ArticleServer, comment *core.CommentServer) *CommunityService {
	return &CommunityService{
		ctx:         ctx,
		plateServ:   plate,
		articleServ: article,
		commentServ: comment,
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

	uid := ctx.Get("uid").(string)
	article.WithAuthor(uid)

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

	if err := cs.articleServ.ListArticle(ctx.Request().Context(), &list); err != nil {
		lg.Errorf("HandleListArticle|ListArticle err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "ListArticle Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cs *CommunityService) HandleCreateComment(ctx echo.Context) error {
	var comment core.Comment
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&comment); err != nil {
		lg.Errorf("HandleCreateComment|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	uid := ctx.Get("uid").(string)
	comment.WithAuthor(uid)
	if err := cs.commentServ.CreateComment(ctx.Request().Context(), &comment); err != nil {
		lg.Errorf("HandleCreateComment|CreateComment err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "CreateComment Error",
		})
	}

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, comment)
}

func (cs *CommunityService) HandleListComment(ctx echo.Context) error {
	var list core.ListComment
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&list); err != nil {
		lg.Errorf("HandleListComment|Decode err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.commentServ.GetCommentList(ctx.Request().Context(), &list)
	if err != nil {
		lg.Errorf("HandleListComment|GetCommentList err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "GetCommentList Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, list)
}

func (cs *CommunityService) HandleDeleteComment(ctx echo.Context) error {
	cmid := ctx.Param("cmid")
	if len(cmid) <= 0 {
		lg.Errorf("HandleDeleteComment|Param Invalid")
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpParamsError, echo.Map{
			"msg": "Params Invalid",
		})
	}

	err := cs.commentServ.DeleteComment(ctx.Request().Context(), cmid)
	if err != nil {
		lg.Errorf("HandleDeleteComment|DeleteComment err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "DeleteComment Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, echo.Map{
		"msg": "DeleteComment Succ",
	})
}

func (cs *CommunityService) HandleGetPlateInfo(ctx echo.Context) error {
	pid := ctx.Param("pid")
	info, err := cs.plateServ.QueryPlateInfo(ctx.Request().Context(), pid)
	if err != nil {
		lg.Errorf("HandleGetPlateInfo|QueryPlateInfo err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryPlateInfo Error",
		})
	}
	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}

func (cs *CommunityService) HandleQueryArticleInfo(ctx echo.Context) error {
	aid := ctx.Param("aid")
	info, err := cs.articleServ.QueryArticleInfo(ctx.Request().Context(), aid)
	if err != nil {
		lg.Errorf("HandleQueryArticleInfo|QueryArticleInfo err:%v", err)
		return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpInternalError, echo.Map{
			"msg": "QueryArticleInfo Error",
		})
	}

	lg.Infof("HandleQueryArticleInfo|QueryArticleInfo|Succ|%s", common.ToStringWithoutError(info))

	return httpx.JsonResponse(ctx, httpx.HttpStatusCode.HttpOK, info)
}
