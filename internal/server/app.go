package server

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"sync"

	"github.com/atopos31/go-veilink/internal/common"
	"github.com/atopos31/go-veilink/internal/config"
	"github.com/google/uuid"
)

type App struct {
	configPath  string
	lock        sync.Mutex
	config      *config.ServerConfig
	listenerMgr *ListenerMgr
	gateway     *Gateway
}

func (a *App) GetClientTunnel(clientID string, tunnelID string) (*config.Listener, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, client := range a.config.Clients {
		if client.ClientID == clientID {
			for _, tunnel := range client.Listeners {
				if tunnel.Uuid == tunnelID {
					return tunnel, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("tunnel not found")
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
			listener.Uuid = uuid.New().String()
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
		Listeners: []*config.Listener{},
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.config.Clients = append(a.config.Clients, &newClient)
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
	a.config.Clients = slices.DeleteFunc(a.config.Clients, func(c *config.Client) bool {
		return c.ClientID == clientID
	})
	return nil
}

func (a *App) GetKey(clientID string) (string, error) {
	key, err := a.listenerMgr.keymap.Get(clientID)
	if err != nil {
		return "", err
	}
	return common.KeyByteToString(key), nil
}

func (a *App) GetClient(clientID string) (*config.Client, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, client := range a.config.Clients {
		if client.ClientID == clientID {
			return client, nil
		}
	}
	return nil, fmt.Errorf("client: %s not found", clientID)
}

func (a *App) GetClientTunnels(clientID string) ([]*config.Listener, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, client := range a.config.Clients {
		if client.ClientID == clientID {
			return client.Listeners, nil
		}
	}
	return nil, fmt.Errorf("client: %s not found", clientID)
}

func (a *App) GetOnline(clientID string) (bool, error) {
	ok := a.gateway.IsOnline(clientID)
	return ok, nil
}

func (a *App) AddClientTunnel(clientID string, tunnel config.Listener) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, client := range a.config.Clients {
		if client.ClientID == clientID {
			tunnel.Uuid = uuid.New().String()
			if err := a.listenerMgr.AddListener(clientID, &tunnel); err != nil {
				return err
			}
			client.Listeners = append(client.Listeners, &tunnel)
			return nil
		}
	}
	return fmt.Errorf("client: %s not found", clientID)
}

func (a *App) RemoveClientTunnel(clientID string, tunnelID string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	if err := a.listenerMgr.RemoveListener(clientID, tunnelID); err != nil {
		return err
	}
	for _, client := range a.config.Clients {
		if client.ClientID == clientID {
			client.Listeners = slices.DeleteFunc(client.Listeners, func(t *config.Listener) bool {
				return t.Uuid == tunnelID
			})
			break
		}
	}
	return nil
}

func (a *App) SaveConfig() error {
	yaml, err := a.config.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(a.configPath, yaml, fs.ModeAppend)
}
