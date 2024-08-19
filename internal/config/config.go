package config

import (
	"github.com/spf13/viper"
)

type ClientConfig struct {
	ServerIp   string `mapstructure:"server_ip"`
	ServerPort int    `mapstructure:"server_port"`
	ClientID   string `mapstructure:"client_id"`
}

func NewClientConfig(configPath string) ClientConfig {
	confViper := viper.New()
	
	confViper.SetConfigFile(configPath)
	var config ClientConfig
	if err := confViper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := confViper.Unmarshal(&config); err != nil {
		panic(err)
	}
	return config
}

type ServerConfig struct {
	Gateway         Gateway    `mapstructure:"gateway"`
	ListenerConfigs []Listener `mapstructure:"listeners"`
}

type Gateway struct {
	Ip   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

type Listener struct {
	ClientID         string `mapstructure:"client_id"`
	PublicProtocol   string `mapstructure:"public_protocol"`
	PublicIP         string `mapstructure:"public_ip"`
	PublicPort       uint16 `mapstructure:"public_port"`
	InternalProtocol string `mapstructure:"internal_protocol"`
	InternalIP       string `mapstructure:"internal_ip"`
	InternalPort     uint16 `mapstructure:"internal_port"`
}

func NewServerConfig(configPath string) ServerConfig {
	confViper := viper.New()

	confViper.SetConfigFile(configPath)
	var config ServerConfig
	if err := confViper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := confViper.Unmarshal(&config); err != nil {
		panic(err)
	}
	return config
}
