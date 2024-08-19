package main

import (
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
	
	config := config.NewClientConfig("../../internal/config/client.toml")
	client := client.NewClient(config)
	client.Run()
}
