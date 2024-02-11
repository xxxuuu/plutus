package app

import (
	"context"

	log "github.com/sirupsen/logrus"
)

var (
	services map[string]Service = make(map[string]Service)
)

type Service interface {
	Name() string
	Init(*Config, *Status, *log.Entry) error
	Run(context.Context) error
}

func RegisterService(s Service) {
	services[s.Name()] = s
}
