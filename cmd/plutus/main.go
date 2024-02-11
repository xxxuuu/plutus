package main

import (
	"context"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"plutus/pkg/app"
	_ "plutus/pkg/notice"
	_ "plutus/pkg/service"
)

func main() {
	log := log.New()

	var config app.Config
	if err := app.LoadConfig("", &config); err != nil {
		log.Error(err)
		return
	}

	app := app.NewApp(
		"Plutus",
		&config,
		log,
	)

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	if err := app.Run(ctx); err != nil {
		log.Error(err)
	}
}
