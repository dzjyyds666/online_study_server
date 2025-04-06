package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func reverseProxy(target string) echo.HandlerFunc {
	return func(c echo.Context) error {
		targetURL, err := url.Parse(target)
		if err != nil {
			return err
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 处理 Host 头，防止后端拒绝请求
		c.Request().Header.Set("X-Forwarded-Host", c.Request().Host)
		c.Request().Host = targetURL.Host

		// 重要：确保 POST 请求的 Body 被正确传输
		if c.Request().Method == http.MethodPost {
			c.Request().Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
		}

		//打印日志信息
		log.Debugf("请求信息：%s", c.Request().URL.String())
		// 代理请求
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func main() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	// 代理不同的微服务
	userGroup := e.Group("/v1/user")
	userGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return reverseProxy("http://127.0.0.1:19001")
	})

	cosGroup := e.Group("/v1/cos")
	cosGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return reverseProxy("http://127.0.0.1:19002")
	})

	classGroup := e.Group("/v1/class")
	classGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return reverseProxy("http://127.0.0.1:19003")
	})

	// 监听端口
	e.Logger.Fatal(e.Start(":19000"))
}
