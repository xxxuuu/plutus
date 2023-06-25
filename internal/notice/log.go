package notice

import "plutus/internal/app"

type LogNotice struct {
}

func (l LogNotice) Notice(ctx app.EventContext, srv any) {
	if log, ok := srv.(LogSender); ok {
		if str, ok := ctx.Value(app.NoticeContent).(string); ok {
			content, logger := log.Log(ctx, str)
			logger.Info(content)
		}
	}
}

type LogSender interface {
	Log(ctx app.EventContext, content string) (string, *app.Logger)
}

func init() {
	l := LogNotice{}
	app.RegisterNotice(l)
}
