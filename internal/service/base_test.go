package service

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"
)

func getClient() (*ethclient.Client, error) {
	nodeAddr := "ws://127.0.0.1:8545"
	client, err := ethclient.Dial(nodeAddr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type baseTestSuite struct {
	suite.Suite
	client *ethclient.Client
}

func (s *baseTestSuite) SetupTest() {
	client, err := getClient()
	s.NoError(err)
	s.client = client
}
