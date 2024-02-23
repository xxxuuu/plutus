package app

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	CacheSize     int                      `koanf:"cache_size"`
	NodeAddress   string                   `koanf:"node_address"`
	DingtalkToken string                   `koanf:"dingtalk_token"`
	BscScanToken  string                   `koanf:"bscscan_token"`
	Services      map[string]ServiceConfig `koanf:"services"`
}

type ServiceConfig struct {
	Enabled bool `koanf:"enabled"`
}

func readConfig() *koanf.Koanf {
	k := koanf.New(".")
	_ = k.Load(file.Provider("config.yaml"), yaml.Parser())
	_ = k.Load(file.Provider("conf/config.yaml"), yaml.Parser())
	return k
}

func LoadConfig[T any](prefix string, cfg *T) error {
	k := readConfig()
	if err := k.Unmarshal(prefix, cfg); err != nil {
		return err
	}
	return nil
}

func LoadServiceConfig[T any](srvName string, cfg *T) error {
	return LoadConfig[T](fmt.Sprintf("services.%s.config", srvName), cfg)
}
