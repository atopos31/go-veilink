package server

import (
	"io/fs"
	"os"
	"slices"
	"sync"

	"github.com/atopos31/go-veilink/internal/common"
	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
)

type App struct {
	configPath  string
	lock        sync.Mutex
	config      *config.ServerConfig
	listenerMgr *ListenerMgr
	gateway     *Gateway
}

func NewApp(configPath string) *App {
	config := config.NewServerConfig(configPath)
	common.InitLogrus(config.LogLevel)

	sessionMgr := NewSessionManager()
	udpSessionMgr := NewUDPSessionManage()
	keymap := NewKeyMap()
	listenerMgr := NewListenerMgr(sessionMgr, udpSessionMgr, keymap)
	gw := NewGateway(config.Gateway, listenerMgr, sessionMgr)
	return &App{configPath: configPath, lock: sync.Mutex{}, config: config, listenerMgr: listenerMgr, gateway: gw}
}

func (a *App) Config() *config.ServerConfig {
	return a.config
}

func (a *App) Start() error {
	if err := a.gateway.Run(); err != nil {
		return err
	}
	for _, client := range a.config.Clients {
		if err := a.listenerMgr.AddClient(client.ClientID); err != nil {
			return err
		}
		for _, listener := range client.Listeners {
			if err := a.listenerMgr.AddListener(client.ClientID, listener); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) AddClient(clientID string) error {
	newClient := config.Client{
		ClientID:  clientID,
		Listeners: []config.Listener{},
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.config.Clients = append(a.config.Clients, newClient)
	if err := a.listenerMgr.AddClient(clientID); err != nil {
		return err
	}
	return nil
}

func (a *App) RemoveClient(clientID string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if err := a.listenerMgr.RemoveClient(clientID); err != nil {
		return err
	}
	a.config.Clients = slices.DeleteFunc(a.config.Clients, func(c config.Client) bool {
		return c.ClientID == clientID
	})
	return nil
}

func (a *App) GetKey(clientID string) (string, error) {
	key, err := a.listenerMgr.keymap.Get(clientID)
	if err != nil {
		return "", err
	}
	return pkg.KeyByteToString(key), nil
}

func (a *App) SaveConfig() error {
	listenerConfigs, err := a.config.Marshal()
	if err != nil {
		return err
	}
	if err := os.WriteFile(a.configPath, listenerConfigs, fs.ModeAppend); err != nil {
		return err
	}
	logrus.Infof("config saved to %s", a.configPath)
	return nil
}
