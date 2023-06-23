package app

const NoticeContent = "content"

type Notice interface {
	// Notice send notice, srv is Service, "any" is used to avoid circular references
	Notice(srv any, content map[string]any)
}

