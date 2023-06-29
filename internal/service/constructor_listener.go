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
	// contract address -> byte code
	byteCodes map[string]string
	// contract address -> contract name
	contractName map[string]string
	retry bool
}

type ConstructorConfig struct {
	// contract name -> contract addresses
	Contracts map[string][]string `koanf:"contracts"`
}

func NewConstructorListener() *ConstructorListener {
	c := &ConstructorListener{
		cfg: &ConstructorConfig{},
		byteCodes: map[string]string{},
		contractName: map[string]string{},
		retry: false,
	}
	return c
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

func (c *ConstructorListener) getByteCode(contract string) (string, error) {
	if byteCode, has := c.appStatus.Cache.Get(fmt.Sprintf("BYTECODE_%s", contract)); has {
		return byteCode.(string), nil
	}

	byteCode, err := c.appStatus.Client.CodeAt(context.Background(), common.HexToAddress(contract), nil)
	if err != nil {
		return "", err
	}
	c.appStatus.Cache.Add(fmt.Sprintf("BYTECODE_%s", contract), string(byteCode))
	return string(byteCode), nil
}

func (c *ConstructorListener) PreRun() {
	c.contractName = map[string]string{}
	c.byteCodes = map[string]string{}
	for name, contracts := range c.cfg.Contracts {
		for i := range contracts {
			contract := contracts[i]
			byteCode, err := c.getByteCode(contract)
			if err != nil {
				c.appStatus.Log.Warnf("%s PreRun(): contract %s get byteCode failed: %s", c.Name(), contract, err)
				continue
			}
			c.contractName[contract] = name
			c.byteCodes[contract] = string(byteCode)
		}
	} 
}

func (c *ConstructorListener) Retry() bool {
	ret := c.retry
	c.retry = false
	return ret
}

func (c *ConstructorListener) NeedHandle(ctx app.EventContext) (bool, error) {
	event := ctx.Event()
	contract := event.Topics[1].Hex()
	byteCode, err := c.getByteCode(contract)
	if err != nil {
		c.appStatus.Log.Errorf("get %s bytecode failed: %s", common.HexToAddress(contract), err)
		c.retry = true
		return false, err
	}

	for i := range c.byteCodes {
		if string(byteCode) == c.byteCodes[i] {
			ctx.Set("Contract", contract)
			ctx.Set("SrcContract", i)
			ctx.Set("SrcContractName", c.contractName[i])
			return true, nil
		}
	}
	return false, nil
}

func (c *ConstructorListener) Execute(ctx app.EventContext) error {
	contract := ctx.Value("Contract").(string)
	srcContract := ctx.Value("SrcContract").(string)
	srcContractName := ctx.Value("SrcContractName").(string)

	ctx.Set(app.NoticeContent,
		fmt.Sprintf("合约地址 %s，与 %s(%s) 相同",
		common.HexToAddress(contract), srcContract, srcContractName))

	c.operator.BroadCast(ctx, c)

	return nil
}

func (c *ConstructorListener) SendDingtalk(ctx app.EventContext) (string, string) {
	token := c.appCfg.DingtalkToken
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
	return token, fmt.Sprintf(json, ctx.Value(app.NoticeContent).(string))
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
