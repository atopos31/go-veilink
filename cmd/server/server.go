package main

import (
	"fmt"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/internal/server"
	"github.com/sirupsen/logrus"
)

func main() {
	config := config.NewServerConfig("../../internal/config/server.toml")
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	sessionMgr := server.NewSessionManager(len(config.ListenerConfigs))
	gw := server.NewGateway(config, sessionMgr)

	for _, listenerConfig := range config.ListenerConfigs {
		listener := server.NewListener(&listenerConfig, sessionMgr)
		go func() {
			defer listener.Close()
			if err := listener.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
		logrus.Debug(fmt.Sprintf("server %s:%d %s<=Veilink=>client %s %s:%d %s", 
			listenerConfig.PublicIP, 
			listenerConfig.PublicPort,
			listenerConfig.PublicProtocol, 
			listenerConfig.ClientID, 
			listenerConfig.InternalIP, 
			listenerConfig.InternalPort,
			listenerConfig.InternalProtocol,
		))
	}

	if err := gw.Run(); err != nil {
		panic(err)
	}
	
}
