package config

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ServerIp   string `mapstructure:"server_ip"`
	ServerPort int    `mapstructure:"server_port"`
	ClientID   string `mapstructure:"client_id"`
	LogLevel   string `mapstructure:"level"`
	Encrypt    bool   `mapstructure:"encrypt"`
	Key        string `mapstructure:"tcp_key"`
}

type ServerConfig struct {
	LogLevel string   `mapstructure:"level"`
	WebUI    WebUI    `mapstructure:"webui"`
	Gateway  Gateway  `mapstructure:"gateway"`
	Clients  []Client `mapstructure:"clients"`
}

type WebUI struct {
	AccessKey string `mapstructure:"access_key"`
	Port      int    `mapstructure:"port"`
	IP        string `mapstructure:"ip"`
}

type Gateway struct {
	Ip        string `mapstructure:"ip"`
	Port      int    `mapstructure:"port"`
	DebugInfo bool   `mapstructure:"debug_info"`
}

type Client struct {
	ClientID  string     `mapstructure:"client_id"`
	Listeners []Listener `mapstructure:"listeners"`
}

type Listener struct {
	ClientID       string `mapstructure:"client_id"`
	Encrypt        bool   `mapstructure:"encrypt"`
	DebugInfo      bool   `mapstructure:"debug_info"`
	PublicProtocol string `mapstructure:"public_protocol"`
	PublicIP       string `mapstructure:"public_ip"`
	PublicPort     uint16 `mapstructure:"public_port"`
	InternalIP     string `mapstructure:"internal_ip"`
	InternalPort   uint16 `mapstructure:"internal_port"`
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
