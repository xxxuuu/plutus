package service

import (
	"testing"

	"plutus/pkg/app"

	"github.com/nanmu42/etherscan-api"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

func TestTransferListener(t *testing.T) {
	suite.Run(t, new(TransferListenerTestSuite))
}

type TransferListenerTestSuite struct {
	baseTestSuite

	wallets []string
}

func (s *TransferListenerTestSuite) SetupTest() {
	s.baseTestSuite.SetupTest()

	s.wallets = []string{"0x95D4ce9cF81AB59E3607Db4de46BD576A4ED3018"}

	s.srv = NewTransferListener()
	s.srv.(*TransferListener).srvCfg = &TransferConfig{
		ThresholdValue: "2000",
		Wallets:        s.wallets,
	}
	s.srv.Init(&app.Config{}, &app.Status{
		Client:        s.client,
		BscScanClient: s.bscscanClient,
	}, logrus.NewEntry(logrus.New()))
}

func (s *TransferListenerTestSuite) TestHandle() {
	s.patch.ApplyMethodFunc(s.bscscanClient, "ERC20Transfers",
		func(contractAddress, address *string, _, _ *int, _, _ int, _ bool) ([]etherscan.ERC20Transfer, error) {
			s.Nil(contractAddress)
			s.Equal("0x7A4B173e6Af66cD7a4312a7AE900222f591F403D", *address)
			return []etherscan.ERC20Transfer{
				{
					TokenSymbol: "BUSDT",
				}, {
					TokenSymbol:     "TEST",
					ContractAddress: "0x9624393cba121b81695b6c3d8ffc9337fe581897",
				}, {
					ContractAddress: "0xfdcca677e59c138fda21055057f57c3f9adf7656",
					TokenSymbol:     "UЅDТ",
				},
			}, nil
		})
	s.ReplayBlockWithRun(36094489)

	s.NotNil(s.noticeMsg)
	transferMsg := s.noticeMsg.(*TransferMsg)
	s.Equal("0x50c647dcb6f7d9e724f342ff5ddc8047f90caee74ca4d083c2b86dbcf4911ade", transferMsg.txHash)
	s.Equal("0x7A4B173e6Af66cD7a4312a7AE900222f591F403D", transferMsg.from)
	s.Equal(s.wallets[0], transferMsg.to)
	s.Equal("6300.00", transferMsg.amount)
	s.Equal(map[string]string{
		"0x9624393cba121b81695b6c3d8ffc9337fe581897": "TEST",
	}, transferMsg.relevantTokens)
}
