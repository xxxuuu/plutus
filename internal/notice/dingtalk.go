package notice

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"plutus/internal/app"
)

const DingtalkUrl = "https://oapi.dingtalk.com/robot/send?access_token=%s"

type Dingtalk struct {
}

func (d Dingtalk) Notice(srv any, content map[string]any) {
	if sender, ok := srv.(DingtalkSender); ok {
		content, token := sender.SendDingtalk(content)
		resp, err := http.Post(fmt.Sprintf(DingtalkUrl, token), "application/json", bytes.NewBuffer([]byte(content)))
		if err != nil {
			log.Println(err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()
	} 
}

type DingtalkSender interface {
	// SendDingtalk return (token, content json string)
	SendDingtalk(content map[string]any) (string, string)
}

func init() {
	d := Dingtalk{}
	app.RegisterNotice(d)
}

