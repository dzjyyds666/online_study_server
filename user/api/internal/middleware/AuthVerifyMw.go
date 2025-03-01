package middleware

import (
	"encoding/json"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/labstack/echo"
	"net/http"
	"user/config"
)

type Token struct {
	Uid  string `json:"uid"`
	Role string `json:"role"`
}

func AuthVerifyMw(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get(httpx.CustomHttpHeader.Authorization.String())

		jwtToken, err := sdk.ParseJwtToken(*config.GloableConfig.Jwt.Secretkey, auth)
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|ParseJwtToken err:%v", err)
			return c.JSON(http.StatusUnauthorized, httpx.HttpResponse{
				StatusCode: httpx.HttpStatusCode.HttpUnauthorized,
				Msg:        "invalid_token", // token非法
			})
		} else {
			var token Token
			err := json.Unmarshal([]byte(jwtToken), &token)
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Token err:%v", err)
				return c.JSON(http.StatusUnauthorized, httpx.HttpResponse{
					StatusCode: httpx.HttpStatusCode.HttpUnauthorized,
					Msg:        "invalid_token", // token非法
				})
			} else {
				c.Set("uid", token.Uid)
				c.Set("role", token.Role)
				return next(c)
			}
		}
	}
}
