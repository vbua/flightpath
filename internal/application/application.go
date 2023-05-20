package application

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/vbua/flightpath/internal/api"
	"github.com/vbua/flightpath/internal/config"
	"github.com/vbua/flightpath/internal/service"
	"go.uber.org/zap"
)

type Application struct {
	cfg     *config.Config
	service *service.FlightPath
	logger  *zap.SugaredLogger

	errChan chan error

	wg sync.WaitGroup
}

func New() *Application {
	return &Application{
		errChan: make(chan error),
	}
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.initLogger(); err != nil {
		return fmt.Errorf("can't init logger: %w", err)
	}

	if err := a.initConfig(); err != nil {
		return fmt.Errorf("can't init config: %w", err)
	}

	a.service = service.NewFlightPath()

	if err := a.initAPIServer(ctx); err != nil {
		return fmt.Errorf("can't init api server: %w", err)
	}

	return nil
}

func (a *Application) initConfig() error {
	var err error

	a.cfg, err = config.ParseConfig()
	if err != nil {
		return fmt.Errorf("can't parse config: %w", err)
	}

	return nil
}

func (a *Application) initAPIServer(ctx context.Context) error {
	apiServer := api.NewRouter(
		a.service,
		a.cfg.ServerOpts,
		a.logger,
	)

	listener, err := net.Listen("tcp4", a.cfg.MainPort)
	if err != nil {
		return fmt.Errorf("can't listen tcp port %s: %w", a.cfg.MainPort, err)
	}

	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		err := apiServer.Start(listener)
		if err != nil {
			a.errChan <- fmt.Errorf("can't start http server: %w", err)
		}
	}()

	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		<-ctx.Done()

		if err := apiServer.Shutdown(); err != nil {
			a.errChan <- fmt.Errorf("can't shutdown http server: %w", err)
		}
	}()

	return nil
}

func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {
	var appErr error

	errWg := sync.WaitGroup{}

	errWg.Add(1)

	go func() {
		defer errWg.Done()

		for err := range a.errChan {
			cancel()
			a.logger.Error(err)
			appErr = err
		}
	}()

	<-ctx.Done()
	a.wg.Wait()
	close(a.errChan)
	errWg.Wait()

	return appErr
}

func (a *Application) initLogger() error {
	log, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("can't create logger: %w", err)
	}

	a.logger = log.Sugar()

	return nil
}
