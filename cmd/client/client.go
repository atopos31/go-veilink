package main

import (
	"flag"

	"github.com/atopos31/go-veilink/internal/client"
	"github.com/atopos31/go-veilink/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	config := config.ClientConfig{}
	flag.StringVar(&config.Key, "key", "", "TCP key")
	flag.StringVar(&config.ServerIp, "ip", "", "Server IP")
	flag.IntVar(&config.ServerPort, "port", 0, "Server Port")
	flag.StringVar(&config.ClientID, "id", "", "Client ID")
	flag.BoolVar(&config.Encrypt, "encrypt", false, "Encrypt")

	flag.Parse()
	client := client.NewClient(config)
	logrus.Infof("Client started %v", config)
	client.Run()
}
