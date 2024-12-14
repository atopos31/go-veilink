package config

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ServerIp   string `mapstructure:"server_ip" yaml:"server_ip"`
	ServerPort int    `mapstructure:"server_port" yaml:"server_port"`
	ClientID   string `mapstructure:"client_id" yaml:"client_id"`
	LogLevel   string `mapstructure:"level" yaml:"level"`
	Encrypt    bool   `mapstructure:"encrypt" yaml:"encrypt"`
	Key        string `mapstructure:"tcp_key" yaml:"tcp_key"`
}

type ServerConfig struct {
	LogLevel string    `mapstructure:"level" yaml:"level"`
	WebUI    WebUI     `mapstructure:"webui" yaml:"webui"`
	Gateway  Gateway   `mapstructure:"gateway" yaml:"gateway"`
	Clients  []*Client `mapstructure:"clients" yaml:"clients"`
}

type WebUI struct {
	AccessKey string `mapstructure:"access_key" yaml:"access_key"`
	Port      int    `mapstructure:"port" yaml:"port"`
	IP        string `mapstructure:"ip" yaml:"ip"`
}

type Gateway struct {
	Ip        string `mapstructure:"ip" yaml:"ip"`
	Port      int    `mapstructure:"port" yaml:"port"`
	DebugInfo bool   `mapstructure:"debug_info" yaml:"debug_info"`
}

type Client struct {
	ClientID  string      `mapstructure:"client_id" yaml:"client_id" json:"client_id"`
	Listeners []*Listener `mapstructure:"listeners" yaml:"listeners" json:"listeners"`
}

type Listener struct {
	Uuid           string `yaml:"-" json:"uuid"` // only use for webui
	ClientID       string `mapstructure:"client_id" yaml:"client_id" json:"client_id"`
	Encrypt        bool   `mapstructure:"encrypt" yaml:"encrypt" json:"encrypt"`
	DebugInfo      bool   `mapstructure:"debug_info" yaml:"debug_info" json:"debug_info"`
	PublicProtocol string `mapstructure:"public_protocol" yaml:"public_protocol" json:"public_protocol"`
	PublicIP       string `mapstructure:"public_ip" yaml:"public_ip" json:"public_ip"`
	PublicPort     uint16 `mapstructure:"public_port" yaml:"public_port" json:"public_port"`
	InternalIP     string `mapstructure:"internal_ip" yaml:"internal_ip" json:"internal_ip"`
	InternalPort   uint16 `mapstructure:"internal_port" yaml:"internal_port" json:"internal_port"`
}

func NewServerConfig(configPath string) *ServerConfig {
	confViper := viper.New()

	confViper.SetConfigFile(configPath)
	var config ServerConfig
	if err := confViper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := confViper.Unmarshal(&config); err != nil {
		panic(err)
	}
	return &config
}

func (c *ServerConfig) Marshal() ([]byte, error) {
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return yaml, nil
}
