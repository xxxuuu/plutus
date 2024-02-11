package service

import (
	log "github.com/sirupsen/logrus"

	"plutus/pkg/app"
	"plutus/pkg/notice"
)

type BaseService struct {
	*app.Status
	log *log.Entry
	cfg *app.Config
}

func (b *BaseService) BroadCast(msg string) {
	log.Infof("broadcast: %s", msg)
	notice.BroadCast(msg, b)
}
