package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/vbua/flightpath/internal/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	app := application.New()

	err := app.Start(ctx)
	if err != nil {
		log.Fatalf("can't start application: %s", err)
	}

	err = app.Wait(ctx, cancel)
	if err != nil {
		log.Fatalf("All systems closed with errors. LastError: %s", err)
	}

	log.Println("All systems closed without errors")
}
