package handler

import (
	"net/http"

	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/logic"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/svc"
	"github.com/dzjyyds666/online_study_server/object_storage/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ApplyUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApplyUploadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewApplyUploadLogic(r.Context(), svcCtx)
		resp, err := l.ApplyUpload(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
