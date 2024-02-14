package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type App struct {
	name     string
	config   *Config
	services map[string]Service
	log      *log.Logger

	*Status
}

type Status struct {
	Client bind.ContractBackend
}

func (app *App) runService(ctx context.Context, srv Service) {
	log := app.log.WithField("service", srv.Name())
	for {
		log.Info("Service running...")
		err := srv.Run(ctx)
		if err != nil {
			log.Errorf("Run failed: %s", err)
			time.Sleep(time.Second)
		}
	}
}

func (app *App) Run(ctx context.Context) error {
	log := app.log.
		WithField("service", app.services).
		WithField("config", app.config)

	client, err := ethclient.Dial(app.config.NodeAddress)
	if err != nil {
		return fmt.Errorf("connect to Node failed: %w", err)
	}
	defer client.Close()
	app.Client = &Client{
		client,
		lru.NewCache[string, any](int(app.config.CacheSize)),
	}

	for _, s := range app.services {
		srvLog := app.log.WithField("service", s.Name())
		err := s.Init(app.config, app.Status, srvLog)
		if err != nil {
			return fmt.Errorf("init service %s failed: %w", s.Name(), err)
		}
	}
	for _, srv := range app.services {
		go app.runService(ctx, srv)
	}

	log.Info("App running...")
	<-ctx.Done()

	return nil
}

func NewApp(name string, config *Config, log *log.Logger) *App {
	srvMap := make(map[string]Service)
	for name, srvCfg := range config.Services {
		if !srvCfg.Enabled {
			continue
		}
		_, exist := services[name]
		if exist {
			srvMap[name] = services[name]
		}
	}

	status := &Status{}
	app := &App{
		name,
		config,
		srvMap,
		log,
		status,
	}

	return app
}
