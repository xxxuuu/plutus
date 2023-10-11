package notice

import (
	"encoding/json"
	"os"
	"plutus/internal/app"
	"strings"
	"testing"
)

type MockDingtalkSender struct{}

func (m *MockDingtalkSender) SendDingtalk(ctx app.EventContext) (string, string) {
	token := os.Getenv("DINGTALK_TOKEN")
	content := `{
	  "msgtype": "markdown",
	  "markdown": {
		"title": "unit testing...",
		"text": "unit testing..."
	  },
	  "at": {
		"atMobiles": [],
		"atUserIds": [],
		"isAtAll": false
	  }
	}`

	return token, content
}

func TestJsonDecode(t *testing.T) {
	jsonStr := `
	{
	"errcode": 1234,
	"errmsg": "invalid parameters"
	}
	`

	var data map[string]interface{}
	_ = json.NewDecoder(strings.NewReader(jsonStr)).Decode(&data)
	if data["errmsg"].(string) != "invalid parameters" {
		t.Errorf("expect: invalid parameters; got: %s", data["errmsg"].(string))
	}
}

func TestNotice(t *testing.T) {
	d := Dingtalk{}
	m := &MockDingtalkSender{}

	ctx := app.NewEventContext(nil)
	if err := d.Notice(ctx, m); err != nil {
		t.Error(err)
	}
}
