package main

import (
	"flag"

	"github.com/atopos31/go-veilink/internal/socks5"
	"github.com/sirupsen/logrus"
)

func main() {
	var ip string
	var port int
	var levelstr string

	flag.StringVar(&ip, "ip", "localhost", "IP address to listen on")
	flag.IntVar(&port, "port", 1080, "Port to listen on")
	flag.StringVar(&levelstr, "level", "debug", "Log level")
	flag.Parse()

	level, err := logrus.ParseLevel(levelstr)
	if err != nil {
		panic(err)
	}

	logrus.SetLevel(level)

	server := socks5.SOCKS5Server{
		IP:   ip,
		Port: port,
	}
	if err := server.Run(); err != nil {
		panic(err)
	}
}
