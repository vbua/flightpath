package api

import (
	"fmt"
	"net"
	"time"

	"github.com/fasthttp/router"
	"github.com/lab259/cors"
	"github.com/valyala/fasthttp"
	"github.com/vbua/flightpath/internal/config"
	"go.uber.org/zap"
)

type Router struct {
	router *router.Router
	serv   *fasthttp.Server
	logger *zap.SugaredLogger
}

func NewMainRouter(cfg config.ServerOpts, logger *zap.SugaredLogger) *Router {
	innerRouter := router.New()
	innerHandler := innerRouter.Handler

	mainRouter := &Router{
		router: innerRouter,
		serv: &fasthttp.Server{
			Handler:            cors.AllowAll().Handler(innerHandler),
			ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
			MaxRequestBodySize: cfg.MaxRequestBodySizeMb * 1024 * 1024,
		},
		logger: logger,
	}

	mainRouter.router.POST("/calculate", mainRouter.calculate)

	return mainRouter
}

func (r *Router) calculate(_ *fasthttp.RequestCtx) {
}

func (r *Router) Start(listener net.Listener) error {
	if err := r.serv.Serve(listener); err != nil {
		return fmt.Errorf("main router listen: %w", err)
	}

	return nil
}

func (r *Router) Shutdown() error {
	if err := r.serv.Shutdown(); err != nil {
		return fmt.Errorf("can't shutdown fasthttp server: %w", err)
	}

	return nil
}
