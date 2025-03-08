package mymiddleware

import (
	"github.com/dzjyyds666/opensource/httpx"
	"github.com/dzjyyds666/opensource/logx"
	"github.com/labstack/echo"
	"github.com/redis/go-redis/v9"
)

type token struct {
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
		//_ := c.Request().Header.Get(httpx.CustomHttpHeader.Authorization.String())

		var auth token
		// rpc调用user服务解析token，返回uid，role

		//判断是否有权限操作
		if auth.Role < permission {
			logx.GetLogger("OS_Server").Errorf("AuthVerifyMw|Permission Denied|%v", auth.Role)
			return httpx.JsonResponse(c, httpx.HttpStatusCode.HttpUnauthorized, echo.Map{
				"msg": "permission denied",
			})
		}

		return next(c)
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
