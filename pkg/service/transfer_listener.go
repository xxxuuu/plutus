package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"plutus/pkg/app"
	"plutus/pkg/common/address"
	"plutus/pkg/common/book"
	"plutus/pkg/common/util"
)

const (
	USDTDecimal = 18
	WBNBDecimal = 18
)

type TransferListener struct {
	BaseService
	srvCfg      *TransferConfig
	bnb         *book.Erc20
	usdt        *book.Erc20
	pancakeSwap *book.PancakeRouterV2
}

type TransferConfig struct {
	Wallets        []string `koanf:"wallets"`
	ThresholdValue string   `koanf:"threshold_value"`
}

func (t *TransferListener) Name() string {
	return "transfer"
}

func (t *TransferListener) Run(ctx context.Context) error {
	var wallets []common.Address
	for _, w := range t.srvCfg.Wallets {
		wallets = append(wallets, common.HexToAddress(w))
	}

	bnbSink := make(chan *book.Erc20Transfer)
	bnbSub, err := t.bnb.WatchTransfer(nil, bnbSink, []common.Address{}, wallets)
	if err != nil {
		return err
	}
	defer bnbSub.Unsubscribe()

	usdtSink := make(chan *book.Erc20Transfer)
	usdtSub, err := t.usdt.WatchTransfer(nil, usdtSink, []common.Address{}, wallets)
	if err != nil {
		return err
	}
	defer usdtSub.Unsubscribe()

	for {
		var event *book.Erc20Transfer
		select {
		case <-ctx.Done():
			return nil
		case event = <-usdtSink:
		case event = <-bnbSink:
		}

		if event == nil {
			continue
		}
		err := t.handle(event)
		if err != nil {
			t.log.WithField("tx hash", event.Raw.TxHash).Errorf("handle failed: %s", err)
		}
	}
}

func (t *TransferListener) handle(event *book.Erc20Transfer) error {
	value := event.Tokens
	if event.Raw.Address != common.HexToAddress(address.USDT) {
		amountOut, err := t.pancakeSwap.GetAmountOut(nil, event.Tokens,
			event.Raw.Address.Big(), common.HexToAddress(address.USDT).Big())
		if err != nil {
			return err
		}
		value = amountOut
	}
	if util.ToDecimal(value, USDTDecimal).LessThan(util.ToDecimal(t.srvCfg.ThresholdValue, 0)) {
		return nil
	}

	t.BroadCast(fmt.Sprintf(
		"[%s] Received transfer event - Tx Hash: %s, From: %s, To: %s, Value: %s USDT\n",
		time.Now().Format("2006-01-02 15:04:05"),
		event.Raw.TxHash.Hex(),
		event.From.Hex(),
		event.To.Hex(),
		event.Tokens,
	))

	return nil
}

func (t *TransferListener) Init(config *app.Config, status *app.Status, log *log.Entry) error {
	t.cfg = config
	t.Status = status
	t.log = log

	err := app.LoadServiceConfig(t.Name(), t.srvCfg)
	if err != nil {
		return err
	}

	bnb, err := book.NewErc20(common.HexToAddress(address.WBNB), t.Client)
	if err != nil {
		return err
	}
	t.bnb = bnb

	usdt, err := book.NewErc20(common.HexToAddress(address.USDT), t.Client)
	if err != nil {
		return err
	}
	t.usdt = usdt

	pancakeSwap, err := book.NewPancakeRouterV2(common.HexToAddress(address.PancakeRouterV2), t.Client)
	if err != nil {
		return err
	}
	t.pancakeSwap = pancakeSwap

	t.log.WithField("config", t.srvCfg).Info("Inited")
	return nil
}

func NewTransferListener() *TransferListener {
	return &TransferListener{}
}

func init() {
	app.RegisterService(NewTransferListener())
}
