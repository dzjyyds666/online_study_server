package middleware

import (
	"github.com/dzjyyds666/online_study_server/user/api/internal/svc"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"net/http"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(httpx.CustomHttpHeader.Authorization.String())

		logx.GetLogger("OS_Server").Infof("JwtConfig|%s", common.ToStringWithoutError(svc.JwtConfig))

		// 解析token
		jwtToken, err := sdk.ParseJwtToken(svc.JwtConfig.SecretKey, token)
		if nil != err {
			logx.GetLogger("OS_Server").Errorf("AuthMiddleware|ParseJwtToken err:%v", err)
			w.Write([]byte("token is invalid"))
			return
		} else {
			r.Header.Set(httpx.CustomHttpHeader.Authorization.String(), jwtToken)
			next(w, r)
		}
	}
}
