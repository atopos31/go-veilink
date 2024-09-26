package main

import (
	"github.com/atopos31/go-veilink/internal/socks5"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	server := socks5.SOCKS5Server{
		IP:   "localhost",
		Port: 1080,
	}
	if err := server.Run(); err != nil {
		panic(err)
	}
}
