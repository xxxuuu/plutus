package app

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)


type Operator struct {
	status *Status
}

func (o *Operator) ByteCode(addr common.Address) ([]byte, error) {
	return o.status.Client.CodeAt(context.Background(), addr, common.Big0)
}

func (o *Operator) BroadCast(ctx EventContext, srv any) {
	for _, n := range notices {
		n.Notice(ctx, srv)	
	}
}
