package api

//go:generate mockgen -destination=router_mock.go -source=router.go -package=api

import (
	"fmt"
	"net"
	"time"

	"github.com/fasthttp/router"
	jsonIter "github.com/json-iterator/go"
	"github.com/lab259/cors"
	"github.com/valyala/fasthttp"
	"github.com/vbua/flightpath/internal/config"
	"github.com/vbua/flightpath/internal/models"
	"go.uber.org/zap"
)

type FlightService interface {
	FindStartAndEndOfPath(flights [][]string) models.Flight
}

type Router struct {
	srv    FlightService
	router *router.Router
	serv   *fasthttp.Server
	logger *zap.SugaredLogger
}

func NewRouter(srv FlightService, cfg config.ServerOpts, logger *zap.SugaredLogger) *Router {
	innerRouter := router.New()
	innerHandler := innerRouter.Handler

	mainRouter := &Router{
		srv:    srv,
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

func (r *Router) calculate(ctx *fasthttp.RequestCtx) {
	var flightPath models.FlightPathRequest

	if err := jsonIter.Unmarshal(ctx.PostBody(), &flightPath); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(fmt.Sprintf("can't unmarshal request: %s", err)))

		return
	}

	path := r.srv.FindStartAndEndOfPath(flightPath.Flights)

	rawBody, err := jsonIter.Marshal(path)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte(fmt.Sprintf("can't marshal result: %s", err)))

		return
	}

	ctx.SetBody(rawBody)
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
