package main

import (
	"embed"
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "path to config file")
	flag.Parse()
	if configPath == "" {
		panic("config path is required")
	}
	config := config.NewServerConfig(configPath)
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	listenerCount := len(config.ListenerConfigs)
	sessionMgr := server.NewSessionManager(listenerCount)
	udpSessionMgr := server.NewUDPSessionManage()
	keymap := server.NewKeyMap()
	gw := server.NewGateway(config, sessionMgr)

	for _, listenerConfig := range config.ListenerConfigs {
		listener := server.NewListener(&listenerConfig, keymap, sessionMgr, udpSessionMgr)
		go func() {
			defer listener.Close()
			if err := listener.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
		if listenerConfig.DebugInfo {
			go listener.DebugInfoTicker(5 * time.Second)
		}
	}

	if config.Gateway.DebugInfo {
		go gw.DebugInfoTicker(5 * time.Second)
	}

	go webServer()

	if err := gw.Run(); err != nil {
		panic(err)
	}
}

//go:embed web/*.html
var htmlFiles embed.FS

//go:embed web/*.css web/*.ico
var cssFiles embed.FS

func webServer() {
	r := gin.Default()
	httpFS := http.FS(htmlFiles)
	cssFS := http.FS(cssFiles)
	r.GET("/login", func(ctx *gin.Context) {
		ctx.FileFromFS("/web/login.html", httpFS)
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.FileFromFS("/web/home.html", httpFS)
	})

	r.StaticFS("/static", cssFS)

	r.GET("/access", access)

	r.Run(":9529")
}

func access(ctx *gin.Context) {
	accessKey := ctx.Query("ak")
	testak := "123456"
	if strings.EqualFold(accessKey, testak) {
		ctx.SetCookie("ak", testak, 3600, "/", "localhost", false, false)
		ctx.String(http.StatusOK, "login success")
	} else {
		ctx.String(http.StatusUnauthorized, "invalid access key")
	}
}
