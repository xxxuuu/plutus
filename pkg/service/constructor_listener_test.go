package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"plutus/pkg/app"
)

func TestConstructorLisenter(t *testing.T) {
	suite.Run(t, new(ConstructorListenerTestSuite))
}

type ConstructorListenerTestSuite struct {
	baseTestSuite

	group  string
	tokenA string
	tokenB string
	tokenC string
}

func (s *ConstructorListenerTestSuite) SetupTest() {
	s.baseTestSuite.SetupTest()

	s.group = "testgroup"
	s.tokenA = "0x5aef33e31e8ab838570652a5a96e5f2fa7aa9b15"
	s.tokenB = "0xb69ba38f8d37a773ef111e627dc1202933ede855"
	s.tokenC = "0x504149ba33a5acb13afa49a2dd0cbbea84edccb5"

	s.srv = NewConstructorListener()
	s.srv.(*ConstructorListener).srvCfg.Tokens = map[string][]string{
		s.group: {s.tokenA},
	}
	s.srv.Init(&app.Config{}, &app.Status{
		Client: s.client,
	}, logrus.NewEntry(logrus.New()))
}

func (s *ConstructorListenerTestSuite) TestConstructorByteCode() {
	byteCodeA, err := s.client.CodeAt(context.Background(), common.HexToAddress(s.tokenA), nil)
	s.NoError(err)

	byteCodeB, err := s.client.CodeAt(context.Background(), common.HexToAddress(s.tokenB), nil)
	s.NoError(err)

	byteCodeC, err := s.client.CodeAt(context.Background(), common.HexToAddress(s.tokenC), nil)
	s.NoError(err)

	s.Equal(byteCodeA, byteCodeB)
	s.Equal(byteCodeA, byteCodeC)
}

func (s *ConstructorListenerTestSuite) TestHandle() {
	blockHeight := 29013912
	s.ReplayBlockWithRun(int64(blockHeight))

	expectedTxHash := "0xe03b736c80e11028d54cfa34261e22bf49fe0385ef1a2f3d3ed1e88dbd3df7c5"
	s.NotNil(s.noticeMsg)
	s.Contains(s.noticeMsg.String(), fmt.Sprintf("%d", blockHeight))
	s.Contains(strings.ToLower(s.noticeMsg.String()), strings.ToLower(s.tokenC))
	s.Contains(s.noticeMsg.String(), s.group)
	s.Contains(s.noticeMsg.String(), expectedTxHash)
}
