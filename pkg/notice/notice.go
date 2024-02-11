package notice

const KeyNoticeContent = "content"

var notices []Notice

type Notice interface {
	Notice(msg string, srv any) error
}

func RegisterNotice(n Notice) {
	notices = append(notices, n)
}

func BroadCast(msg string, srv any) {
	for _, n := range notices {
		n.Notice(msg, srv)
	}
}
