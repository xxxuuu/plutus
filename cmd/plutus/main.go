package main

import (
	"go.uber.org/zap"

	"plutus/internal/app"
	_ "plutus/internal/notice"
	_ "plutus/internal/service"
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
		&app.Logger{SugaredLogger: *sugar},
	)

	app.Run()
}
