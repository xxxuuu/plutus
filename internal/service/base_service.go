package service

import "plutus/internal/app"

type baseService struct {
	*app.Status
	cfg      *app.Config
	operator app.Operator
}

func (b baseService) PreRun() {
}

func (b baseService) Log(ctx app.EventContext, content string) (string, *app.Logger) {
	return content, b.Logger
}
