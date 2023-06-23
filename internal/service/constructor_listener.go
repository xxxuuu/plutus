package service

import (
	"context"
	"fmt"
	"plutus/internal/app"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

const PancakeSwapAddress = "0xca143ce32fe78f1f7019d7d551a6402fc5350c73"
const PancakeSwapPairCreated = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"

type ConstructorListener struct {
	baseService
	cfg *ConstructorConfig
	byteCodes []string
}

type ConstructorConfig struct {
	Contracts []string `koanf:"contracts"`
}

func NewConstructorListener() *ConstructorListener {
	return &ConstructorListener{
	}
}

func (c *ConstructorListener) Name() string {
	return "ConstructorListener"
}

func (c *ConstructorListener) EthFilter() ethereum.FilterQuery {
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(PancakeSwapAddress)},
		Topics: [][]common.Hash{
			{common.HexToHash(PancakeSwapPairCreated)},
			{}, {},
		},
	}
	return filter
}

func (c *ConstructorListener) NeedHandle(event *app.Event) bool {
	byteCode, err := c.operator.ByteCode(event.Address)
	if err != nil {
		c.appStatus.Log.Errorf("get %s bytecode failed: %s", event.Address.String(), err)
		return false
	}
	for i := range c.byteCodes {
		if string(byteCode) == c.byteCodes[i] {
			return true
		}
	}
	return false
}

func (c *ConstructorListener) PreRun() {
	for _, contract := range c.cfg.Contracts {
		byteCode, err := c.appStatus.Client.CodeAt(context.Background(), common.HexToAddress(contract), nil)
		if err != nil {
			c.appStatus.Log.Warnf("%s PreRun(): contract %s get byteCode failed: %s", c.Name(), contract, err)
			continue
		}
		c.byteCodes = append(c.byteCodes, string(byteCode))
	}
}

func (c *ConstructorListener) Execute(event *app.Event) {
	c.operator.BroadCast(c, map[string]any{
		app.NoticeContent: fmt.Sprintf("contract %s discovered", event.Address.String()),
		"contract": event.Topics[0].Hex(),
	})
}

func (c *ConstructorListener) SendDingtalk(content map[string]any) (string, string) {
	token := c.appCfg.DingtalkToken
	json := `{
	  "msgtype": "markdown",
	  "markdown": {
		"title": "上链检测",
		"text": "contract address: %s"
	  },
	  "at": {
		"atMobiles": [],
		"atUserIds": [],
		"isAtAll": false
	  }
	}`
	contractAddr := content["contract"].(string)
	return token, fmt.Sprintf(json, contractAddr)
}

func (c *ConstructorListener) Init(config *app.Config, status *app.Status, operator app.Operator) {
	c.appCfg = config
	c.appStatus = status
	c.operator = operator
	app.LoadConfig("constructor", c.cfg)
	c.appStatus.Log.Infof("%s loaded config %v", c.Name(), c.cfg)
}

func init() {
	app.RegisterService(NewConstructorListener())
}
