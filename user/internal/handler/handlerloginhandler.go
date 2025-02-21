package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github/dzjyyds666/online_study_server/user/internal/logic"
	"github/dzjyyds666/online_study_server/user/internal/svc"
	"github/dzjyyds666/online_study_server/user/internal/types"
)

func HandlerLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewHandlerLoginLogic(r.Context(), svcCtx)
		resp, err := l.HandlerLogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
