package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/atopos31/go-veilink/internal/handler"
	"github.com/atopos31/go-veilink/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "path to config file")
	flag.Parse()
	if strings.EqualFold(configPath, "") {
		panic("config path is required")
	}

	app := server.NewApp(configPath)
	if err := app.Start(); err != nil {
		panic(err)
	}

	handler := handler.NewServerHandler(app)
	addr := fmt.Sprintf("%s:%d", app.Config().WebUI.IP, app.Config().WebUI.Port)
	r := webServer(handler)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("shutting down web server")
	if err := app.SaveConfig(); err != nil {
		logrus.Errorf("failed to save config %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("failed to shutdown web server %v", err)
	}
}

//go:embed web/*.html
var htmlFiles embed.FS

//go:embed web/*.css web/*.ico
var staticFiles embed.FS

func webServer(handler *handler.ServerHandler) http.Handler {
	r := gin.Default()
	htmlFS, err := fs.Sub(htmlFiles, "web")
	if err != nil {
		panic(err)
	}

	staticFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		panic(err)
	}

	r.GET("/login", func(ctx *gin.Context) {
		ctx.FileFromFS("login.html", http.FS(htmlFS))
	})

	r.GET("/", handler.Auth, func(ctx *gin.Context) {
		ctx.FileFromFS("home.html", http.FS(htmlFS))
	})

	r.StaticFS("/static", http.FS(staticFS))

	api := r.Group("/api")
	api.GET("/access", handler.Access)

	clients := api.Group("/clients", handler.Auth)
	clients.GET("/", handler.GetClients)
	clients.POST("/:clientID", handler.AddClient)
	clients.GET("/:clientID/online", handler.GetClientOnline)
	clients.DELETE("/:clientID", handler.RemoveClient)
	clients.GET("/:clientID/key", handler.GetClientKey)
	clients.GET("/:clientID/tunnels", handler.GetClientTunnels)
	clients.POST("/:clientID/tunnels", handler.AddClientTunnel)
	clients.DELETE("/:clientID/tunnels/:tunnelID", handler.RemoveClientTunnel)
	clients.PUT("/:clientID/tunnels/:tunnelID", handler.UpdateClientTunnel)
	return r
}
