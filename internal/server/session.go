package server

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/smux"
)

var (
	ErrNotConnected   = errors.New("not connected")
	ErrClientIsOnline = errors.New("client is online")
)

type UDPsession struct {
	RemoteAddr string
	LocalAddr  string
	tunnelConn pkg.VeilConn
}

type UDPSessionManage struct {
	sessionMu sync.Mutex
	sessions  map[string]*UDPsession
}

func NewUDPSessionManage() *UDPSessionManage {
	return &UDPSessionManage{
		sessions: make(map[string]*UDPsession),
	}
}

func (usm *UDPSessionManage) Get(key string) (*UDPsession, error) {
	usm.sessionMu.Lock()
	defer usm.sessionMu.Unlock()

	session, ok := usm.sessions[key]
	if ok {
		return session, nil
	} else {
		return nil, errors.New("not found")
	}
}

func (usm *UDPSessionManage) Add(key string, session *UDPsession) {
	usm.sessionMu.Lock()
	defer usm.sessionMu.Unlock()

	usm.sessions[key] = session
	go usm.CleanCache(key)
}

func (usm *UDPSessionManage) CleanCache(key string) {
	tick := time.NewTicker(time.Second * 20)
	defer tick.Stop()
	for range tick.C {
		usm.sessionMu.Lock()
		usm.sessions[key].tunnelConn.Close()
		delete(usm.sessions, key)
		usm.sessionMu.Unlock()
		break
	}
}

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

func (sm *SessionManager) GetSessionConnByID(clientID string) (pkg.VeilConn, error) {
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
	logrus.Debugf("client %s start online", clientID)
	<-sm.sessions[clientID].Connection.CloseChan()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, clientID)
	logrus.Debugf("client %s is offline", clientID)
}
