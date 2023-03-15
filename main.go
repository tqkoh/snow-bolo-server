package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/tqkoh/snow-bolo-server/game"
	"github.com/tqkoh/snow-bolo-server/streamer"
	"github.com/tqkoh/snow-bolo-server/utils"
)

func main() {
	s := streamer.NewStreamer()

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Logger.SetLevel(log.DEBUG)
	e.Logger.SetHeader("${time_rfc3339} ${prefix} ${short_file} ${line} |")
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: "${time_rfc3339} method = ${method} | uri = ${uri} | code = ${status} ${error}\n"}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "front: https://tqk.blue/snow-bolo/")
	})

	api := e.Group("/api")
	{
		api.GET("/ping", func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		})
		api.GET("/ws", func(c echo.Context) error {
			s.ConnectWS(c, func(c *streamer.Client) {
				game.ProcessDeadDisconnected(s, c.Id)
				utils.Del(s.Clients, c.Id)
			})
			return nil
		})
	}

	go game.GameLoop(s)
	go s.Listen(game.HandlerWebSocket)

	e.Logger.Panic(e.Start(":3939"))
}
