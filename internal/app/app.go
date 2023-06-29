package app

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/lru"
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
	failoverCh chan failoverReq
	operator Operator
	log *Logger

	status *Status
}

type Status struct {
	Log *Logger
	Client *ethclient.Client
	Cache *lru.Cache[string, any]
}

type Logger struct {
	zap.SugaredLogger
}

type Event struct {
	types.Log
}

type Filter interface {
	EthFilter() ethereum.FilterQuery
	NeedHandle(EventContext) (bool, error)
}

type Executor interface {
	PreRun()
	Execute(EventContext) error
}

type Service interface {
	Name() string
	Filter
	Executor
	Retry() bool
	Init(*Config, *Status, Operator)
}

type failoverReq struct {
	serviceName string
	event EventContext
}

func EmptyStatus() *Status {
	return &Status {
		nil,
		nil,
		lru.NewCache[string, any](256),
	}
}

func failoverWithoutEvent(srvName string) failoverReq {
	return failoverReq{
		serviceName: srvName,
		event: NewEventContext(nil),
	}
}

func failoverWithEvent(srvName string, event EventContext) failoverReq {
	return failoverReq{
		serviceName: srvName,
		event: event,
	}
}

func (app *App) serviceFailover() {
	go func()  {
		for req := range app.failoverCh {
			service := app.services[req.serviceName]

			retryEvent := &req.event
			if !service.Retry() {
				retryEvent = nil
			}

			// TODO: limiting
			app.executeService(service, retryEvent)
		}	
	}()	
}

func (app *App) executeService(srv Service, retryEvent *EventContext) {
	go func() {
		filter := srv.EthFilter()
		logCh := make(chan types.Log)

		sub, err := app.status.Client.SubscribeFilterLogs(context.Background(), filter, logCh)
		if err != nil {
			app.log.Warnf("Service %s subscribe failed: %s", srv.Name(), err)
			app.failoverCh<-failoverWithoutEvent(srv.Name())
			return
		}

		srv.PreRun()
		app.log.Infof("Service %s running...", srv.Name())

		handleEvent := func(ctx EventContext) {
			needhandle, err := srv.NeedHandle(ctx)
			if err != nil {
				app.log.Errorf("Service %s NeedHandle() failed: %s", srv.Name(), err)
				app.failoverCh<-failoverWithEvent(srv.Name(), ctx)
			}
			if !needhandle {
				return
			}
			if err := srv.Execute(ctx); err != nil {
				app.log.Errorf("Service %s Execute() failed: %s", srv.Name(), err)
				app.failoverCh<-failoverWithEvent(srv.Name(), ctx)
			}
		}

		if retryEvent != nil {
			handleEvent(*retryEvent)
		}

		for {
			select {
			case err := <-sub.Err():
				app.status.Log.Warnf("Service %s error: %s, restart...", srv.Name(), err)
				app.failoverCh<-failoverWithoutEvent(srv.Name())
				return
			case log := <-logCh:
				ctx := NewEventContext(&Event{log})
				handleEvent(ctx)
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
		app.executeService(srv, nil)
	}

	app.status.Log.Info("app running...")

	<-app.exitCh
}

func NewApp(name string, config *Config, options []Option, logger *Logger) *App {
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
		make(chan failoverReq),
		Operator{status},
		logger,
		status,
	}

	for _, opt := range options {
		opt(app)
	}

	app.status.Log.Infof("app config %v", app.config)

	for _, s := range app.services {
		s.Init(app.config, app.status, app.operator)
	}

	return app
}
