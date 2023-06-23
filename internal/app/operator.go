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

func (o *Operator) BroadCast(srv any, content map[string]any) {
	for _, n := range o.status.notices {
		n.Notice(srv, content)	
	}
}
