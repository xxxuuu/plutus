package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"plutus/internal/app"
)

const DingtalkUrl = "https://oapi.dingtalk.com/robot/send?access_token=%s"

type Dingtalk struct {
}

func (d Dingtalk) Notice(ctx app.EventContext, srv any) error {
	if sender, ok := srv.(DingtalkSender); ok {
		token, content := sender.SendDingtalk(ctx)
		resp, err := http.Post(fmt.Sprintf(DingtalkUrl, token), "application/json", bytes.NewBuffer([]byte(content)))
		if err != nil {
			return err
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
		if data["errcode"].(float64) != 0 {
			return errors.New(data["errmsg"].(string))
		}
	} 
	return nil
}

type DingtalkSender interface {
	// SendDingtalk return (token, content json string)
	SendDingtalk(ctx app.EventContext) (string, string)
}

func init() {
	d := Dingtalk{}
	app.RegisterNotice(d)
}

