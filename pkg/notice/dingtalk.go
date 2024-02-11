package notice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const DingtalkUrl = "https://oapi.dingtalk.com/robot/send?access_token=%s"

type Dingtalk struct{}

func (d Dingtalk) Notice(msg string, srv any) error {
	if sender, ok := srv.(DingtalkNotifier); ok {
		token, content := sender.DingtalkMsg(msg)
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

type DingtalkNotifier interface {
	DingtalkMsg(msg string) (token string, content string)
}

func init() {
	RegisterNotice(Dingtalk{})
}
