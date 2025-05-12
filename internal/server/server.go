package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Config struct {
	Dev bool
}

type Server struct {
	e *echo.Echo
}

func NewServer(conf Config) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = conf.Dev

	u := newUI(uiConfig{
		Dev:       conf.Dev,
		AssetPath: "/assets",
	})
	u.ConfigureServer(e)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "dashboard.html", map[string]string{
			"View": "dashboard",
		})
	})

	return &Server{
		e: e,
	}
}

func (s *Server) Run(addr string) error {
	fmt.Printf("Dashi started on %s\n", addr)
	return s.e.Start(addr)
}
