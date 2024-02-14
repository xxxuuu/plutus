package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"plutus/pkg/app"
	"plutus/pkg/common/address"
	"plutus/pkg/common/book"
	"plutus/pkg/notice"
)

type ConstructorListener struct {
	BaseService
	srvCfg *ConstructorConfig
	// token address -> byte code
	byteCodes map[string]string
	// token address -> token group
	tokenGroup map[string]string

	factory *book.PancakeFactoryV2
}

type ConstructorConfig struct {
	// token group -> token addresses
	Tokens map[string][]string `koanf:"tokens"`
}

func NewConstructorListener() *ConstructorListener {
	c := &ConstructorListener{
		srvCfg:     &ConstructorConfig{},
		byteCodes:  map[string]string{},
		tokenGroup: map[string]string{},
	}
	return c
}

func (c *ConstructorListener) Name() string {
	return "constructor"
}

func (c *ConstructorListener) getByteCode(token string) (string, error) {
	byteCode, err := c.Client.CodeAt(context.Background(), common.HexToAddress(token), nil)
	if err != nil {
		return "", err
	}
	return string(byteCode), nil
}

func (c *ConstructorListener) PreRun() {
	c.tokenGroup = map[string]string{}
	c.byteCodes = map[string]string{}
	for group, tokens := range c.srvCfg.Tokens {
		for i := range tokens {
			token := tokens[i]
			byteCode, err := c.getByteCode(token)
			if err != nil {
				c.log.WithField("token", token).Warnf("get byteCode failed: %s", err)
				continue
			}
			c.tokenGroup[token] = group
			c.byteCodes[token] = string(byteCode)
		}
	}
}

func (c *ConstructorListener) WatchEvent(sink chan *book.PancakeFactoryV2PairCreated) (ethereum.Subscription, error) {
	return c.factory.WatchPairCreated(nil, sink, []common.Address{}, []common.Address{})
}

func (c *ConstructorListener) Run(ctx context.Context) error {
	c.PreRun()

	sink := make(chan *book.PancakeFactoryV2PairCreated)
	sub, err := c.WatchEvent(sink)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return err
		case event := <-sink:
			err := c.handle(event)
			if err != nil {
				c.log.WithField("tx hash", event.Raw.TxHash).Errorf("handle failed: %s", err)
			}
		}
	}
}

func (c *ConstructorListener) handle(event *book.PancakeFactoryV2PairCreated) error {
	log := c.log.
		WithField("tx hash", event.Raw.TxHash)

	token0 := event.Token0.Hex()
	byteCode0, err := c.getByteCode(token0)
	if err != nil {
		log.WithField("token0", common.HexToAddress(token0)).Errorf("get bytecode failed: %s", err)
		return err
	}

	token1 := event.Token1.Hex()
	byteCode1, err := c.getByteCode(token1)
	if err != nil {
		log.WithField("token0", common.HexToAddress(token1)).Errorf("get bytecode failed: %s", err)
		return err
	}

	for addr := range c.byteCodes {
		needHandle := false
		token := token0
		if string(byteCode0) == c.byteCodes[addr] {
			needHandle = true
			token = token0
		} else if string(byteCode1) == c.byteCodes[addr] {
			needHandle = true
			token = token1
		}

		if needHandle {
			c.BroadCast(notice.TextMsg(fmt.Sprintf(c.msgTemplate(),
				time.Now().Format(time.DateTime),
				event.Raw.BlockNumber,
				common.HexToAddress(token),
				addr,
				c.tokenGroup[addr],
				event.Raw.TxHash,
			)), c)

			return nil
		}
	}
	return nil
}

func (c *ConstructorListener) msgTemplate() string {
	return `
通知时间: %s

区块高度: %d

合约地址 %s

与 %s(%s) 相同

事件 Hash: %s`
}

func (c *ConstructorListener) DingtalkMsg(msg notice.Msg) (string, string) {
	token := c.cfg.DingtalkToken
	json := `{
	  "msgtype": "markdown",
	  "markdown": {
		"title": "上链检测",
		"text": "%s"
	  },
	  "at": {
		"atMobiles": [],
		"atUserIds": [],
		"isAtAll": false
	  }
	}`
	return token, fmt.Sprintf(json, msg)
}

func (c *ConstructorListener) Init(config *app.Config, status *app.Status, log *log.Entry) error {
	c.cfg = config
	c.log = log
	c.Status = status

	err := app.LoadServiceConfig(c.Name(), &c.srvCfg)
	if err != nil {
		return err
	}

	factory, err := book.NewPancakeFactoryV2(common.HexToAddress(address.PancakeFactoryV2), c.Client)
	if err != nil {
		return err
	}
	c.factory = factory

	c.log.WithField("config", c.srvCfg).Info("Inited")
	return nil
}

func init() {
	app.RegisterService(NewConstructorListener())
}
