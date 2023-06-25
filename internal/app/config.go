package app

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	NodeAddress     string   `koanf:"node_address"`
	DingtalkToken   string   `koanf:"dingtalk_token"`
}

func readConfig() (*koanf.Koanf) {
	k := koanf.New(".")
	_ = k.Load(file.Provider("config.yaml"), yaml.Parser())
	_ = k.Load(file.Provider("conf/config.yaml"), yaml.Parser())
	// _ = k.Load(env.ProviderWithValue(prefix, ".", func(s string, v string) (string, interface{}) {
	// 	key := strings.ToLower(strings.TrimPrefix(s, prefix))
	// 	if strings.Contains(v, ",") {
	// 		return key, strings.Split(v, ",")
	// 	}
	// 	return key, v
	// }), nil)
	return k
}

func LoadConfig[T any](prefix string, cfg *T) error {
	k := readConfig()
	if err := k.Unmarshal(prefix, cfg); err != nil {
		return err
	}
	return nil
}
