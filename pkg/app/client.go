package app

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	RetryCnt = 5

	CacheKeyCodeAt = "CodeAt_%s"
)

type Client interface {
	ethereum.ChainReader
	ethereum.TransactionReader
	ethereum.ChainStateReader
	ethereum.FeeHistoryReader
	ethereum.BlockNumberReader
	ethereum.ChainIDReader
	bind.ContractBackend
	Close()
}

type CachedClient struct {
	Client
	cache *lru.Cache[string, any]
}

func NewCachedClient(baseClient Client, cacheSize int) *CachedClient {
	return &CachedClient{
		Client: baseClient,
		cache:  lru.NewCache[string, any](cacheSize),
	}
}

func (c *CachedClient) Close() {
	c.Client.Close()
	c.cache.Purge()
}

func (c *CachedClient) retryWithBackOff(maxRetryCnt int, fn func() error) error {
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

func (c *CachedClient) CodeAt(ctx context.Context, contract ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
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

type SimulatedClient struct {
	Client
	blockNumber *big.Int
	logSub      []*SimulatedSubscription[types.Log]
	blockSub    []*SimulatedSubscription[*types.Header]
}

type SimulatedSubscription[T types.Log | *types.Header] struct {
	ethereum.Subscription
	query  ethereum.FilterQuery
	ch     chan<- T
	active bool
}

func (s *SimulatedSubscription[T]) Unsubscribe() {
	s.Subscription.Unsubscribe()
	s.active = false
}

func NewSimulatedClient(client Client, blockNumber *big.Int) *SimulatedClient {
	return &SimulatedClient{
		Client:      client,
		blockNumber: blockNumber,
		logSub:      []*SimulatedSubscription[types.Log]{},
		blockSub:    []*SimulatedSubscription[*types.Header]{},
	}
}

func (c *SimulatedClient) Close() {
	c.Client.Close()
	c.logSub = nil
	c.blockSub = nil
}

func (c *SimulatedClient) SetBlockNumber(blockNumber *big.Int) {
	c.blockNumber = blockNumber
}

func (c *SimulatedClient) calcBlockNumber(blockNumber *big.Int) *big.Int {
	if blockNumber == nil || blockNumber.Cmp(c.blockNumber) > 0 {
		return c.blockNumber
	}
	return blockNumber
}

func (c *SimulatedClient) FetchNewBlock() error {
	c.blockNumber.Add(c.blockNumber, big.NewInt(1))
	// get the block at the height from mainnet, and publish it
	header, err := c.Client.HeaderByNumber(context.Background(), c.blockNumber)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for i := range c.blockSub {
		sub := c.blockSub[i]
		if !sub.active {
			continue
		}
		wg.Add(1)
		go func() {
			sub.ch <- header
			wg.Done()
		}()
	}
	for i := range c.logSub {
		sub := c.logSub[i]
		if !sub.active {
			continue
		}
		query := sub.query
		query.FromBlock = c.blockNumber
		query.ToBlock = c.blockNumber
		logs, err := c.Client.FilterLogs(context.Background(), query)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			for _, log := range logs {
				sub.ch <- log
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func (c *SimulatedClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return c.Client.HeaderByNumber(ctx, c.calcBlockNumber(number))
}

func (c *SimulatedClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return c.Client.BlockByNumber(ctx, c.calcBlockNumber(number))
}

func (c *SimulatedClient) CodeAt(ctx context.Context, contract ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
	return c.Client.CodeAt(ctx, contract, c.calcBlockNumber(blockNumber))
}

func (c *SimulatedClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	if query.ToBlock != nil && query.ToBlock.Cmp(c.blockNumber) > 0 {
		query.ToBlock = c.blockNumber
	}
	sub, err := c.Client.SubscribeFilterLogs(ctx, query, ch)
	simSub := &SimulatedSubscription[types.Log]{
		Subscription: sub,
		query:        query,
		ch:           ch,
		active:       true,
	}
	c.logSub = append(c.logSub, simSub)
	return simSub, err
}

func (c *SimulatedClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	sub, err := c.Client.SubscribeNewHead(ctx, ch)
	simSub := &SimulatedSubscription[*types.Header]{
		Subscription: sub,
		ch:           ch,
		active:       true,
	}
	c.blockSub = append(c.blockSub, simSub)
	return simSub, err
}
