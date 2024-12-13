package handler

import (
	"net/http"
	"strings"

	"github.com/atopos31/go-veilink/internal/server"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	app *server.App
}

func NewServerHandler(app *server.App) *ServerHandler {
	return &ServerHandler{app: app}
}

func (s *ServerHandler) Auth(ctx *gin.Context) {
	ak, err := ctx.Cookie("ak")
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	}
	if strings.EqualFold(ak, s.app.Config().WebUI.AccessKey) {
		ctx.Next()
	} else {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	}
}

func (s *ServerHandler) Access(ctx *gin.Context) {
	accessKey := ctx.Query("ak")
	if strings.EqualFold(accessKey, s.app.Config().WebUI.AccessKey) {
		host := ctx.Request.Host
		domain := strings.Split(host, ":")[0]
		ctx.SetCookie("ak", accessKey, 3600, "/", domain, false, false)
		ctx.String(http.StatusOK, "login success")
	} else {
		ctx.String(http.StatusUnauthorized, "invalid access key")
	}
}
