package middleware

import (
	"encoding/json"
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/dzjyyds666/opensource/sdk"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
	"user/api/config"
	"user/api/internal/core"
)

type Token struct {
	Uid  string `json:"uid"`
	Role string `json:"role"`
}

func AuthMw(permission string, ds *redis.Client) echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return AuthVerifyMw(handlerFunc, permission, ds)
	}
}

func AuthVerifyMw(next echo.HandlerFunc, permission string, redis *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get(httpx.CustomHttpHeader.Authorization.String())

		// 先从redis中查询token是否过期和合法
		result, err := redis.Get(c.Request().Context(), core.RedisTokenKey).Result()
		if err != nil || auth != result {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Get Token From Redis Error|%v", err)
			//return c.JSON(http.StatusUnauthorized, httpx.HttpResponse{
			//	StatusCode: httpx.HttpStatusCode.HttpUnauthorized,
			//	Msg:        "invalid_token", // token非法
			//})

			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "invalid_token",
			})
		}

		jwtToken, err := sdk.ParseJwtToken(*config.GloableConfig.Jwt.Secretkey, auth)
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|ParseJwtToken err:%v", err)
			//return c.JSON(http.StatusUnauthorized, httpx.HttpResponse{
			//	StatusCode: httpx.HttpStatusCode.HttpUnauthorized,
			//	Msg:        "invalid_token", // token非法
			//})
			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "invalid_token",
			})
		} else {
			var token Token
			err := json.Unmarshal([]byte(jwtToken), &token)
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Token err:%v", err)
				//return c.JSON(http.StatusUnauthorized, httpx.HttpResponse{
				//	StatusCode: httpx.HttpStatusCode.HttpUnauthorized,
				//	Msg:        "invalid_token", // token非法
				//})
				return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
					"msg": "invalid_token",
				})
			} else {
				c.Set("uid", token.Uid)
				c.Set("role", token.Role)
				return next(c)
			}
		}
	}
}
