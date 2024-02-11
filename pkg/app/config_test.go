package app

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

type ConfigTestSuite struct {
	suite.Suite

	patch *gomonkey.Patches

	rawConfig      string
	expectedConfig *Config
	srv            DummyService
}

func (s *ConfigTestSuite) SetupTest() {
	s.patch = gomonkey.NewPatches()

	s.expectedConfig = &Config{
		CacheSize:     1024,
		NodeAddress:   "wss://bscrpc.com",
		DingtalkToken: "c84b143107be7dfa1add2d67fdfb59c84b143107be7dfa1add2d67fdfb59",
		Services: map[string]ServiceConfig{
			"dummy": {
				Enabled: true,
			},
		},
	}
	s.srv = DummyService{
		&DummyServiceConfig{
			Key1: "value1",
			Key2: "value2",
		},
	}

	log.Infof("%p", readConfig)
	s.rawConfig = fmt.Sprintf(`
cache_size: %d
node_address: %s
dingtalk_token: %s
services:
  %s:
    enabled: %t
    config:
      key1: %s
      key2: %s`,
		s.expectedConfig.CacheSize, s.expectedConfig.NodeAddress, s.expectedConfig.DingtalkToken,
		"dummy", s.expectedConfig.Services["dummy"].Enabled,
		s.srv.config.Key1, s.srv.config.Key2)

	s.patch.ApplyFunc(readConfig, func() *koanf.Koanf {
		k := koanf.New(".")
		err := k.Load(rawbytes.Provider([]byte(s.rawConfig)), yaml.Parser())
		if err != nil {
			log.Error(err)
		}
		return k
	})

	log.Infof("%p", readConfig)
}

func (s *ConfigTestSuite) TearDownTest() {
	s.patch.Reset()
}

func (s *ConfigTestSuite) TestLoadConfig() {
	var actualConfig Config
	s.NoError(LoadConfig("", &actualConfig))
	s.Equal(s.expectedConfig, &actualConfig)
}

func (s *ConfigTestSuite) TestLoadServiceConfig() {
	var actualSrvConfig DummyServiceConfig
	s.NoError(LoadServiceConfig("dummy", &actualSrvConfig))
	s.Equal(s.srv.config, &actualSrvConfig)
}

type DummyService struct {
	config *DummyServiceConfig
}

type DummyServiceConfig struct {
	Key1 string `koanf:"key1"`
	Key2 string `koanf:"key2"`
}
