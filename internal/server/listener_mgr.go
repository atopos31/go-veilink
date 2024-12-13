package server

import (
	"errors"
	"sync"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/pkg"
)

type ListenerMgr struct {
	sessionMgr    *SessionManager
	udpSessionMgr *UDPSessionManage
	keymap        *keymap
	lock          sync.Mutex
	listenersMap  map[string][]*Listener
}

func NewListenerMgr(sessionMgr *SessionManager, udpSessionMgr *UDPSessionManage, keymap *keymap) *ListenerMgr {
	return &ListenerMgr{
		sessionMgr:    sessionMgr,
		udpSessionMgr: udpSessionMgr,
		keymap:        keymap,
		lock:          sync.Mutex{},
		listenersMap:  make(map[string][]*Listener),
	}
}

func (lm *ListenerMgr) AddListener(clientID string, listenerConfig config.Listener) error {
	lm.lock.Lock()
	defer lm.lock.Unlock()
	if _, ok := lm.listenersMap[clientID]; !ok {
		return errors.New("client id not found")
	}
	key, err := lm.keymap.Get(clientID)
	if err != nil {
		return err
	}
	listener := NewListener(&listenerConfig, key, lm.sessionMgr, lm.udpSessionMgr)
	if err := listener.ListenAndServe(); err != nil {
		return err
	}

	lm.listenersMap[clientID] = append(lm.listenersMap[clientID], listener)
	return nil
}

func (lm *ListenerMgr) AddClient(clientID string) error {
	lm.lock.Lock()
	defer lm.lock.Unlock()
	if _, ok := lm.listenersMap[clientID]; ok {
		return errors.New("client id already exists")
	}
	lm.listenersMap[clientID] = make([]*Listener, 0)
	key, err := pkg.GenChacha20Key()
	if err != nil {
		return err
	}
	lm.keymap.Set(clientID, key)
	return nil
}

func (lm *ListenerMgr) RemoveClient(clientID string) error {
	lm.lock.Lock()
	defer lm.lock.Unlock()
	if _, ok := lm.listenersMap[clientID]; !ok {
		return errors.New("client id not found")
	}
	for _, listener := range lm.listenersMap[clientID] {
		listener.Close()
	}
	delete(lm.listenersMap, clientID)
	return nil
}

func (lm *ListenerMgr) CheckExist(clientID string) bool {
	lm.lock.Lock()
	defer lm.lock.Unlock()
	_, ok := lm.listenersMap[clientID]
	return ok
}
