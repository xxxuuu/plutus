package notice

const KeyNoticeContent = "content"

var notices []Notice

type Notice interface {
	Notice(msg Msg, srv any) error
}
type Msg interface {
	String() string
}

type TextMsg string

func (m TextMsg) String() string {
	return string(m)
}

func RegisterNotice(n Notice) {
	notices = append(notices, n)
}

func BroadCast(msg Msg, srv any) {
	for _, n := range notices {
		n.Notice(msg, srv)
	}
}
