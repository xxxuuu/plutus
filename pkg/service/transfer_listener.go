package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"plutus/pkg/app"
	"plutus/pkg/common/address"
	"plutus/pkg/common/book"
	"plutus/pkg/common/util"
	"plutus/pkg/notice"
)

const (
	USDTDecimal = 18
	WBNBDecimal = 18
)

var (
	BlockKeyword = []string{
		"USD",
		"BNB",
		"Cake-LP",
		"claim",
		"rewards",
	}
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

type TransferMsg struct {
	txHash string
	from   string
	to     string
	amount string
	// type: token address -> token name
	relevantTokens map[string]string
}

func (t *TransferMsg) String() string {
	return fmt.Sprintf("[%s] Received transfer event - Tx Hash: %s, From: %s, To: %s, Value: %s USDT, Relevant Tokens: %v",
		time.Now().Format("2006-01-02 15:04:05"),
		t.txHash,
		t.from,
		t.to,
		t.amount,
		t.relevantTokens,
	)
}

func (t *TransferMsg) HumanReadableMsg() string {
	relevantTokens := strings.Builder{}
	for addr, name := range t.relevantTokens {
		relevantTokens.WriteString(fmt.Sprintf("- [%s](%s)\n", name, addr))
	}

	tmpl := `
### Tx Hash
[%s](https://bscscan.com/tx/%s)

### 发款方
[%s](https://www.oklink.com/cn/bsc/address/%s)

### 收款方
[%s](https://www.oklink.com/cn/bsc/address/%s)

### 金额
%s USDT

### 关联币种
%s`
	return fmt.Sprintf(tmpl,
		t.txHash, t.txHash, t.from, t.from, t.to, t.to, t.amount, relevantTokens)
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
		return fmt.Errorf("bnb watch failed: %w", err)
	}
	defer bnbSub.Unsubscribe()

	usdtSink := make(chan *book.Erc20Transfer)
	usdtSub, err := t.usdt.WatchTransfer(nil, usdtSink, []common.Address{}, wallets)
	if err != nil {
		return fmt.Errorf("usdt watch failed: %w", err)
	}
	defer usdtSub.Unsubscribe()

	for {
		var event *book.Erc20Transfer
		select {
		case <-ctx.Done():
			return nil
		case err := <-bnbSub.Err():
			return fmt.Errorf("bnb subscription error: %w", err)
		case err := <-usdtSub.Err():
			return fmt.Errorf("usdt subscription error: %w", err)
		case event = <-usdtSink:
		case event = <-bnbSink:
		}

		if event == nil {
			continue
		}
		err := t.handle(ctx, event)
		if err != nil {
			t.log.WithField("tx hash", event.Raw.TxHash).Errorf("handle failed: %s", err)
		}
	}
}

func (t *TransferListener) handle(ctx context.Context, event *book.Erc20Transfer) error {
	value := event.Tokens
	if event.Raw.Address != common.HexToAddress(address.USDT) {
		amountOut, err := t.pancakeSwap.GetAmountOut(nil, event.Tokens,
			event.Raw.Address.Big(), common.HexToAddress(address.USDT).Big())
		if err != nil {
			return fmt.Errorf("get amount out failed: %w", err)
		}
		value = amountOut
	}
	if util.ToDecimal(value, USDTDecimal).LessThan(util.ToDecimal(t.srvCfg.ThresholdValue, 0)) {
		return nil
	}

	tokens, err := t.relevantTokens(ctx, event.From)
	if err != nil {
		t.log.WithField("eoa", event.From.Hex()).Warnf("get relevant tokens failed: %s", err)
		tokens = map[string]string{}
	}
	t.BroadCast(&TransferMsg{
		txHash:         event.Raw.TxHash.Hex(),
		from:           event.From.Hex(),
		to:             event.To.Hex(),
		amount:         util.ToDecimal(value, USDTDecimal).StringFixed(2),
		relevantTokens: tokens,
	}, t)

	return nil
}

func (t *TransferListener) relevantTokens(ctx context.Context, eoa common.Address) (map[string]string, error) {
	nowBlockHeight, err := t.Client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("get block number failed: %w", err)
	}
	// a month ago
	startBlock := int(nowBlockHeight - 3*20*24*30)
	endBlock := int(nowBlockHeight)
	addr := eoa.Hex()

	txList, err := t.BscScanClient.ERC20Transfers(nil, &addr, &startBlock, &endBlock, 1, 30, true)
	if err != nil {
		return nil, fmt.Errorf("get erc20 transfers failed: %w", err)
	}
	ret := map[string]string{}
	for _, tx := range txList {
		block := false
		for _, word := range BlockKeyword {
			if strings.Contains(strings.ToLower(tx.TokenSymbol), strings.ToLower(word)) {
				block = true
				break
			}
		}
		if !block {
			ret[tx.ContractAddress] = tx.TokenSymbol
		}
	}
	return ret, nil
}

func (t *TransferListener) DingtalkMsg(msg notice.Msg) (token string, content string) {
	transferMsg := msg.(*TransferMsg)
	json := `{
	  "msgtype": "markdown",
	  "markdown": {
		"title": "交易捕获: %s USDT",
		"text": "%s"
	  },
	  "at": {
		"atMobiles": [],
		"atUserIds": [],
		"isAtAll": false
	  }
	}`
	return t.cfg.DingtalkToken, fmt.Sprintf(json, transferMsg.amount, transferMsg.HumanReadableMsg())
}

func (t *TransferListener) Init(config *app.Config, status *app.Status, log *log.Entry) error {
	t.cfg = config
	t.Status = status
	t.log = log

	err := app.LoadServiceConfig(t.Name(), &t.srvCfg)
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
