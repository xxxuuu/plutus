package app

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	RetryCnt = 5

	CacheKeyCodeAt = "CodeAt_%s"
)

type Client struct {
	*ethclient.Client
	cache *lru.Cache[string, any]
}

func (c *Client) retryWithBackOff(maxRetryCnt int, fn func() error) error {
	var err error
	for retry := 0; retry <= maxRetryCnt; retry++ {
		err = fn()
		if err != nil {
			sleepTime := math.Pow(2, float64(retry))
			time.Sleep(time.Duration(sleepTime) * 100 * time.Microsecond)
		}
	}
	return err
}

func (c *Client) CodeAt(ctx context.Context, contract ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
	key := fmt.Sprintf(CacheKeyCodeAt, contract)
	value, ok := c.cache.Get(key)
	if ok {
		return value.([]byte), nil
	}

	var code []byte
	err := c.retryWithBackOff(RetryCnt, func() error {
		_code, err := c.Client.CodeAt(ctx, contract, blockNumber)
		if err != nil {
			return err
		}
		code = _code
		return nil
	})
	if err != nil {
		return nil, err
	}
	return code, nil
}

// func (c *Client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
// 	err := c.retryWithBackOff(RetryCnt, func() error {
// 		s, err := app.Client.SubscribeFilterLogs(context.Background(), filter, logCh)
// 		if err != nil {
// 			log.Warnf("Subscribe failed: %s, retrying...", err)
// 			return err
// 		}
// 		sub = s
// 		return nil
// 	})
// }
