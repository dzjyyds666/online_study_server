package middleware

import (
	"class/api/config"
	"encoding/json"
	"github.com/dzjyyds666/opensource/common"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/labstack/echo"
)

type Token struct {
	Uid  string `json:"uid"`
	Role int    `json:"role"`
}

func AuthMw(permission int) echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return AuthVerifyMw(handlerFunc, permission)
	}
}

func AuthVerifyMw(next echo.HandlerFunc, permission int) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get(httpx.CustomHttpHeader.Authorization.String())
		jwtToken, err := sdk.ParseJwtToken(*config.GloableConfig.Jwt.Secretkey, auth)
		if err != nil {
			logx.GetLogger("study").Errorf("AuthVerifyMw|ParseJwtToken err:%v", err)
			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "invalid_token",
			})
		} else {
			var token Token
			s := (*jwtToken)["data"].(string)
			err = json.Unmarshal([]byte(s), &token)
			if err != nil {
				logx.GetLogger("study").Errorf("AuthVerifyMw|Token err:%v", err)
				return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
					"msg": "invalid_token",
				})
			} else {

				logx.GetLogger("study").Infof("AuthVerifyMw|Token Verify Success|%v", common.ToStringWithoutError(token))
				if token.Role < permission {
					logx.GetLogger("study").Errorf("AuthVerifyMw|permission denied")
					return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
						"msg": "permission denied",
					})
				}
				c.Set("uid", token.Uid)
				c.Set("role", token.Role)
				return next(c)
			}
		}
	}
}

var UserRole = struct {
	Admin   int
	Teacher int
	Student int
}{
	Admin:   3,
	Teacher: 2,
	Student: 1,
}
