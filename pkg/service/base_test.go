package service

import (
	"context"
	"math/big"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nanmu42/etherscan-api"
	"github.com/stretchr/testify/suite"

	"plutus/pkg/app"
	"plutus/pkg/notice"
)

// exec "anvil --fork-url=https://bscrpc.com" before unit tests
const AnvilEndpoint = "ws://127.0.0.1:8545"

func (s *baseTestSuite) getAnvilClient() *ethclient.Client {
	client, err := ethclient.Dial(AnvilEndpoint)
	s.NoError(err)
	return client
}

type baseTestSuite struct {
	suite.Suite

	client        *app.SimulatedClient
	bscscanClient *etherscan.Client
	patch         *gomonkey.Patches
	noticeMsg     notice.Msg
	srv           app.Service
}

func (s *baseTestSuite) SetupTest() {
	s.patch = gomonkey.NewPatches()
	s.patch.ApplyFunc(notice.BroadCast, func(msg notice.Msg, srv any) {
		s.noticeMsg = msg
	})

	s.client = app.NewSimulatedClient(s.getAnvilClient(), nil)
	s.bscscanClient = etherscan.New(etherscan.Mainnet, "")
}

func (s *baseTestSuite) TearDownTest() {
	s.patch.Reset()
	s.noticeMsg = nil
	s.client.Close()
}

func (s *baseTestSuite) ReplayBlockWithRun(blockHeight int64) {
	s.client.SetBlockNumber(big.NewInt(blockHeight - 1))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s.NoError(s.srv.Run(ctx))
	}()

	time.Sleep(1 * time.Second)
	s.NoError(s.client.FetchNewBlock())
	time.Sleep(1 * time.Second)
}
