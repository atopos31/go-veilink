package main

import (
	"embed"
	"flag"
	"net/http"
	"strings"

	"github.com/atopos31/go-veilink/internal/handler"
	"github.com/atopos31/go-veilink/internal/server"
	"github.com/gin-gonic/gin"
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
	defer app.SaveConfig()
	handler := handler.NewServerHandler(app)
	webServer(handler)
}

//go:embed web/*.html
var htmlFiles embed.FS

//go:embed web/*.css web/*.ico
var staticFiles embed.FS

func webServer(handler *handler.ServerHandler) {
	r := gin.Default()
	httpFS := http.FS(htmlFiles)
	staticFS := http.FS(staticFiles)
	r.GET("/login", func(ctx *gin.Context) {
		ctx.FileFromFS("/web/login.html", httpFS)
	})

	r.GET("/", handler.Auth, func(ctx *gin.Context) {
		ctx.FileFromFS("/web/home.html", httpFS)
	})

	r.StaticFS("/static", staticFS)

	api := r.Group("/api")
	api.GET("/access", handler.Access)

	r.Run(":9529")
}
