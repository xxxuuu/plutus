package main

import (
	"plutus/internal/app"

	"go.uber.org/zap"

	_ "plutus/internal/service"
	_ "plutus/internal/notice"
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
		&app.Logger{ *sugar },
	)

	app.Run()
}
