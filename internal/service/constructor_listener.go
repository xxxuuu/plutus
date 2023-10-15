package service

import (
	"context"
	"fmt"
	"plutus/internal/app"
	"plutus/internal/common/address"
	"plutus/internal/common/book"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

const (
	PancakeSwapPairCreated = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
)

type ConstructorListener struct {
	baseService
	srvCfg *ConstructorConfig
	// token address -> byte code
	byteCodes map[string]string
	// token address -> token group
	tokenGroup map[string]string
	retry      bool

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
		retry:      false,
	}
	return c
}

func (c *ConstructorListener) Name() string {
	return "ConstructorListener"
}

func (c *ConstructorListener) EthFilter() ethereum.FilterQuery {
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(address.PancakeFactoryV2)},
		Topics: [][]common.Hash{
			{common.HexToHash(PancakeSwapPairCreated)},
			{},
			{},
		},
	}
	return filter
}

func (c *ConstructorListener) getByteCode(token string) (string, error) {
	if byteCode, has := c.Cache.Get(fmt.Sprintf("BYTECODE_%s", token)); has {
		return byteCode.(string), nil
	}

	byteCode, err := c.Client.CodeAt(context.Background(), common.HexToAddress(token), nil)
	if err != nil {
		return "", err
	}
	c.Cache.Add(fmt.Sprintf("BYTECODE_%s", token), string(byteCode))
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
				c.Logger.Warnf("%s PreRun(): token %s get byteCode failed: %s", c.Name(), token, err)
				continue
			}
			c.tokenGroup[token] = group
			c.byteCodes[token] = string(byteCode)
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

	pairCreated, err := c.factory.PancakeFactoryV2Filterer.ParsePairCreated(event.Log)
	if err != nil {
		return false, err
	}

	token0 := pairCreated.Token0.Hex()
	byteCode0, err := c.getByteCode(token0)
	if err != nil {
		c.Logger.Errorf("get %s bytecode failed: %s", common.HexToAddress(token0), err)
		c.retry = true
		return false, err
	}

	token1 := pairCreated.Token1.Hex()
	byteCode1, err := c.getByteCode(token1)
	if err != nil {
		c.Logger.Errorf("get %s bytecode failed: %s", common.HexToAddress(token1), err)
		c.retry = true
		return false, err
	}

	for i := range c.byteCodes {
		needHandle := false
		token := token0
		if string(byteCode0) == c.byteCodes[i] {
			needHandle = true
			token = token0
		} else if string(byteCode1) == c.byteCodes[i] {
			needHandle = true
			token = token1
		}

		if needHandle {
			ctx.Set("TxHash", event.TxHash.Hex())
			ctx.Set("Token", token)
			ctx.Set("SrcToken", i)
			ctx.Set("SrcTokenGroup", c.tokenGroup[i])
			return true, nil
		}
	}
	return false, nil
}

func (c *ConstructorListener) Execute(ctx app.EventContext) error {
	event := ctx.Event()
	token := ctx.Value("Token").(string)
	srcContract := ctx.Value("SrcToken").(string)
	srcContractName := ctx.Value("SrcTokenGroup").(string)
	txHash := ctx.Value("TxHash").(string)

	ctx.Set(app.NoticeContent,
		fmt.Sprintf(`
通知时间: %s

区块高度: %d

合约地址 %s

与 %s(%s) 相同

事件 Hash: %s`,
			time.Now().Format(time.DateTime),
			event.BlockNumber,
			common.HexToAddress(token),
			srcContract,
			srcContractName,
			txHash,
		),
	)

	c.operator.BroadCast(ctx, c)

	return nil
}

func (c *ConstructorListener) SendDingtalk(ctx app.EventContext) (string, string) {
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
	return token, fmt.Sprintf(json, ctx.Value(app.NoticeContent).(string))
}

func (c *ConstructorListener) Init(config *app.Config, status *app.Status, operator app.Operator) {
	c.cfg = config
	c.Status = status
	c.operator = operator
	factory, err := book.NewPancakeFactoryV2(common.HexToAddress(address.PancakeFactoryV2), c.Client)
	if err != nil {
		panic(err)
	}
	c.factory = factory
	_ = app.LoadConfig("constructor", c.srvCfg)
	c.Logger.Infof("%s loaded config %v", c.Name(), c.srvCfg)
}

func init() {
	app.RegisterService(NewConstructorListener())
}
