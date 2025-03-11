package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/dzjyyds666/opensource/common"
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
	Role int    `json:"role"`
}

func AuthMw(permission int, ds *redis.Client) echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return AuthVerifyMw(handlerFunc, permission, ds)
	}
}

func AuthVerifyMw(next echo.HandlerFunc, permission int, redis *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		uid := c.Param("uid")
		tokenKey := fmt.Sprintf(core.RedisTokenKey, uid)
		auth := c.Request().Header.Get(httpx.CustomHttpHeader.Authorization.String())
		// 先从redis中查询token是否过期和合法
		result, err := redis.Get(c.Request().Context(), tokenKey).Result()
		if err != nil || auth != result {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Get Token From Redis Error|%v", err)
			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "invalid_token",
			})
		}

		jwtToken, err := sdk.ParseJwtToken(*config.GloableConfig.Jwt.Secretkey, auth)
		if err != nil {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|ParseJwtToken err:%v", err)
			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "invalid_token",
			})
		} else {
			var token Token
			s := (*jwtToken)["data"].(string)
			err = json.Unmarshal([]byte(s), &token)
			if err != nil {
				logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Token err:%v", err)
				return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
					"msg": "invalid_token",
				})
			} else {

				logx.GetLogger("OS_Server").Infof("AuthVerifyMw|Token Verify Success|%v", common.ToStringWithoutError(token))
				if token.Role < permission {
					logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|permission denied")
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
