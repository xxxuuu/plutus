package service

import (
	"context"
	"plutus/internal/common/address"
	"plutus/internal/common/book"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

func TestConstructorLisenter(t *testing.T) {
	suite.Run(t, new(ConstructorListenerTestSuite))
}

type ConstructorListenerTestSuite struct {
	baseTestSuite
}

func (s *ConstructorListenerTestSuite) SetupTest() {
	s.baseTestSuite.SetupTest()
}

func (s *ConstructorListenerTestSuite) TestConstructorByteCode() {
	client, err := getClient()
	s.NoError(err)

	byteCodeA, err := client.CodeAt(context.Background(), common.HexToAddress("0x5aef33e31e8ab838570652a5a96e5f2fa7aa9b15"), nil)
	s.NoError(err)

	byteCodeB, err := client.CodeAt(context.Background(), common.HexToAddress("0xb69ba38f8d37a773ef111e627dc1202933ede855"), nil)
	s.NoError(err)

	byteCodeC, err := client.CodeAt(context.Background(), common.HexToAddress("0x504149ba33a5acb13afa49a2dd0cbbea84edccb5"), nil)
	s.NoError(err)

	s.Equal(byteCodeA, byteCodeB)
	s.Equal(byteCodeA, byteCodeC)
}

func (s *ConstructorListenerTestSuite) TestSubscribe() {
	client, err := getClient()
	s.NoError(err)

	factory, err := book.NewPancakeFactoryV2(common.HexToAddress(address.PancakeFactoryV2), client)
	s.NoError(err)

	end := uint64(29013912)
	iter, err := factory.PancakeFactoryV2Filterer.FilterPairCreated(
		&bind.FilterOpts{
			Start:   29013912,
			End:     &end,
			Context: context.Background(),
		},
		[]common.Address{},
		[]common.Address{},
	)
	s.NoError(err)

	s.True(iter.Next())
	s.Equal(
		iter.Event.Token0,
		common.HexToAddress("0x504149ba33a5acb13afa49a2dd0cbbea84edccb5"),
	)
}
