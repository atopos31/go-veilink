package server

import (
	"errors"
	"net"
	"sync"

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
	muxsess, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ClientID:   clientID,
		Connection: muxsess,
	}
	sm.sessions[clientID] = sess
	return sess, nil
}

func (sm *SessionManager) Range(f func(k string, v *Session) bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for k, v := range sm.sessions {
		ok := f(k, v)
		if !ok {
			delete(sm.sessions, k)
		}
	}
}
