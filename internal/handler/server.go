package handler

import (
	"net/http"
	"strings"

	"github.com/atopos31/go-veilink/internal/config"
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

func (s *ServerHandler) GetClients(ctx *gin.Context) {
	clientIDs := make([]string, 0)
	for _, client := range s.app.Config().Clients {
		clientIDs = append(clientIDs, client.ClientID)
	}
	ctx.JSON(http.StatusOK, clientIDs)
}

func (s *ServerHandler) AddClient(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	if err := s.app.AddClient(clientID); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	} else {
		ctx.String(http.StatusOK, "client added")
	}
}

func (s *ServerHandler) RemoveClient(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	if err := s.app.RemoveClient(clientID); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	} else {
		ctx.String(http.StatusOK, "client removed")
	}
}

func (s *ServerHandler) GetClientKey(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	key, err := s.app.GetKey(clientID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.String(http.StatusOK, key)
}

func (s *ServerHandler) GetClientTunnels(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	tunnels, err := s.app.GetClientTunnels(clientID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	ctx.JSON(http.StatusOK, tunnels)
}

func (s *ServerHandler) GetClientOnline(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	online, err := s.app.GetOnline(clientID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	ctx.JSON(http.StatusOK, online)
}

func (s *ServerHandler) GetClientTunnel(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	tunnelID := ctx.Param("tunnelID")
	tunnel, err := s.app.GetClientTunnel(clientID, tunnelID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	ctx.JSON(http.StatusOK, tunnel)
}

func (s *ServerHandler) AddClientTunnel(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	var tunnel config.Listener
	if err := ctx.ShouldBindJSON(&tunnel); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := s.app.AddClientTunnel(clientID, tunnel); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	} else {
		ctx.String(http.StatusOK, "tunnel added")
	}
}

func (s *ServerHandler) RemoveClientTunnel(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	tunnelID := ctx.Param("tunnelID")
	if err := s.app.RemoveClientTunnel(clientID, tunnelID); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
	} else {
		ctx.String(http.StatusOK, "tunnel removed")
	}
}

func (s *ServerHandler) UpdateClientTunnel(ctx *gin.Context) {
	clientID := ctx.Param("clientID")
	tunnelID := ctx.Param("tunnelID")
	var tunnel config.Listener
	if err := ctx.ShouldBindJSON(&tunnel); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	oldTunnel, err := s.app.GetClientTunnel(clientID, tunnelID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.app.RemoveClientTunnel(clientID, tunnelID); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.app.AddClientTunnel(clientID, tunnel); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		s.app.AddClientTunnel(clientID, *oldTunnel)
	} else {
		ctx.String(http.StatusOK, "tunnel updated")
	}

}
