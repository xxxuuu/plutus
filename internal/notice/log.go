package notice

import "plutus/internal/app"

type LogNotice struct {
}

func (l LogNotice) Notice(srv any, content map[string]any) {
	if log, ok := srv.(LogSender); ok {
		if str, ok := content[app.NoticeContent].(string); ok {
			content, logger := log.Log(str)
			logger.Info(content)
		}
	}
}

type LogSender interface {
	Log(content string) (string, *app.Logger)
}

func init() {
	l := LogNotice{}
	app.RegisterNotice(l)
}
