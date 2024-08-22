package server

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xtaci/smux"
)

var (
	ErrNotConnected   = errors.New("not connected")
	ErrClientIsOnline = errors.New("client is online")
)

type Session struct {
	ClientID   string        // 客户端ID
	Connection *smux.Session // 双向连接 server <=> client
}

type SessionManager struct {
	mu       sync.Mutex
	count    int
	sessions map[string]*Session
}

func NewSessionManager(count int) *SessionManager {

	return &SessionManager{
		count:    count,
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) GetSessionConnByID(clientID string) (net.Conn, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess := sm.sessions[clientID]
	if sess == nil {
		return nil, ErrNotConnected
	}
	return sess.Connection.OpenStream()
}

func (sm *SessionManager) AddSession(clientID string, conn net.Conn) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldsess := sm.sessions[clientID]
	if oldsess != nil {
		return nil, ErrClientIsOnline
	}
	smuxconfig := smux.DefaultConfig()
	smuxconfig.KeepAliveInterval = 1 * time.Second
	smuxconfig.KeepAliveTimeout = 2 * time.Second
	muxsess, err := smux.Server(conn, smuxconfig)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ClientID:   clientID,
		Connection: muxsess,
	}
	go sm.CheckAlive(clientID)
	sm.sessions[clientID] = sess
	return sess, nil
}

// 检测到客户端离线后删除session
func (sm *SessionManager) CheckAlive(clientID string) {
	logrus.Debug(fmt.Sprintf("client %s start online", clientID))
	<-sm.sessions[clientID].Connection.CloseChan()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, clientID)
	logrus.Debug(fmt.Sprintf("client %s is offline", clientID))
}

// Debbug online info
func (sm *SessionManager) DebugInfo() {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		sm.mu.Lock()
		logrus.Debug(fmt.Sprintf("↓↓ session is online: %d/%d", len(sm.sessions), sm.count))
		for k := range sm.sessions {
			logrus.Debug(k)
		}
		sm.mu.Unlock()
	}

}
