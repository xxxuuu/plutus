package main

import (
	"plutus/internal/app"

	"go.uber.org/zap"
)

func main() {
	var config app.Config
	if err := app.LoadConfig("", &config); err != nil {
		panic(err)
	}

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	app := app.NewApp(
		"Plutus",
	    &config,
		[]app.Option{},
		[]app.Service{},
		&app.Logger{ *sugar },
	)

	sugar.Info("app running...")
	app.Run()
}
