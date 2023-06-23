package service

import "plutus/internal/app"

type baseService struct {
	appCfg *app.Config
	appStatus *app.Status
	operator app.Operator
}

func (b baseService) PreRun() {

}

func (b baseService) Log(content string) (string, *app.Logger) {
	return content, b.appStatus.Log
}
