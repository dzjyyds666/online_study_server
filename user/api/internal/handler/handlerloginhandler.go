package handler

import (
	"github.com/dzjyyds666/online_study_server/user/api/internal/logic"
	"github.com/dzjyyds666/online_study_server/user/api/internal/svc"
	"github.com/dzjyyds666/online_study_server/user/api/internal/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
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
