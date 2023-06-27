package app

type Operator struct {
	status *Status
}

func (o *Operator) BroadCast(ctx EventContext, srv any) {
	for _, n := range notices {
		n.Notice(ctx, srv)	
	}
}
