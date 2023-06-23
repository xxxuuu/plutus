package app

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type App struct {
	name     string
	config   *Config
	services map[string]Service
	noticeCh map[string]chan Event
	exitCh   chan struct{}
	failoverCh chan string
	operator Operator
	log *Logger

	status *Status
}

type Status struct {
	Log *Logger
	Client *ethclient.Client
	notices []Notice
}

type Logger struct {
	zap.SugaredLogger
}

type Event struct {
	types.Log
}

type Filter interface {
	EthFilter() ethereum.FilterQuery
	NeedHandle(*Event) bool
}

type Executor interface {
	PreRun()
	Execute(*Event)
}

type Service interface {
	Name() string
	Filter
	Executor
	Init(*Config, *Status, Operator)
}

func EmptyStatus() *Status {
	return &Status {
		nil,
		nil,
		[]Notice{},
	}
}

func (app *App) serviceFailover() {
	go func()  {
		for name := range app.failoverCh {
			app.executeService(app.services[name])
		}	
	}()	
}

func (app *App) executeService(srv Service) {
	filter := srv.EthFilter()
	logCh := make(chan types.Log)
	sub, err := app.status.Client.SubscribeFilterLogs(context.Background(), filter, logCh)
	if err != nil {
		app.log.Errorf("Service %s subscribe failed: %s", srv.Name(), err)
		return
	}

	go func() {
		srv.PreRun()
		app.log.Infof("Service %s running...", srv.Name())
		for {
			select {
			case <-sub.Err():
				app.failoverCh<-srv.Name()
				return
			case log := <-logCh:
				event := Event{log}
				if srv.NeedHandle(&event) {
					srv.Execute(&event)
				}
			}
		}
	}()
}

func (app *App) Run() {
	client, err := ethclient.Dial(app.config.NodeAddress)
	if err != nil {
    	app.log.Errorf("connect to Node failed: %s", err)
		return
	}
	app.status.Client = client
	app.serviceFailover()

	for _, srv := range app.services {
		app.executeService(srv)
	}

	<-app.exitCh
}

func NewApp(name string, config *Config, options []Option, services []Service, logger *Logger) *App {
	srvMap := make(map[string]Service)

	for _, srv := range services {
		srvMap[srv.Name()] = srv
	}

	status := EmptyStatus()
	status.Log = logger
	app := &App{
		name,
		config,
		srvMap,
		make(map[string]chan Event),
		make(chan struct{}),
		make(chan string),
		Operator{status},
		logger,
		status,
	}

	for _, opt := range options {
		opt(app)
	}

	for _, s := range app.services {
		s.Init(app.config, app.status, app.operator)
	}

	return app
}
