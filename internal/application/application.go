package application

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Application struct {
	logger *zap.SugaredLogger

	errChan chan error

	wg sync.WaitGroup
}

func New() *Application {
	return &Application{
		errChan: make(chan error),
	}
}

func (a *Application) Start(_ context.Context) error {
	if err := a.initLogger(); err != nil {
		return fmt.Errorf("can't init logger: %w", err)
	}

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
