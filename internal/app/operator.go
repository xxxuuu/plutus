package app

type Operator struct {
	status *Status
}

func (o *Operator) BroadCast(ctx EventContext, srv any) {
	for _, n := range notices {
		if err := n.Notice(ctx, srv); err != nil {
			o.status.Log.Errorf("notice failed: %s", err)
		}
	}
}
