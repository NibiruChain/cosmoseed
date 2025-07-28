package cosmoseed

import (
	"errors"
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type Config struct {
	NodeKeyFile  string `yaml:"nodeKeyFile" default:"node_key.json"`
	AddrBookFile string `yaml:"addrBookFile" default:"addrbook.json"`

	ListenAddr string `yaml:"listenAddr" default:"tcp://0.0.0.0:26656"`
	LogLevel   string `yaml:"logLevel" default:"info"`

	MaxInboundPeers         int  `yaml:"maxInboundPeers" default:"2000"`
	MaxOutboundPeers        int  `yaml:"maxOutboundPeers" default:"20"`
	MaxPacketMsgPayloadSize int  `yaml:"maxPacketMsgPayloadSize" default:"1024"`
	AllowNonRoutable        bool `yaml:"allowNonRoutable"`
	PeerQueueSize           int  `yaml:"peerQueueSize" default:"1000"`
	DialWorkers             int  `yaml:"dialWorkers" default:"20"`

	ChainID string `yaml:"chainID"`
	Seeds   string `yaml:"seeds"`

	ApiAddr string `yaml:"apiAddr" default:"0.0.0.0:8080"`
}

func (cfg *Config) Save(path string) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err = ensurePath(path); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func DefaultConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, defaults.Set(cfg)
}

func ReadConfigFromFile(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error in config file unmarshal: %v", err)
	}
	return &cfg, defaults.Set(&cfg)
}

func (cfg *Config) Validate() error {
	if cfg.ChainID == "" {
		return errors.New("chainID is required")
	}
	return nil
}
