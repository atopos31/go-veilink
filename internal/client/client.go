package client

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/atopos31/go-veilink/internal/common"
	"github.com/atopos31/go-veilink/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/smux"
)

type Client struct {
	serverAddr string
	clientID   string
	key        string
}

func NewClient(conf config.ClientConfig) *Client {
	return &Client{
		serverAddr: fmt.Sprintf("%s:%d", conf.ServerIp, conf.ServerPort),
		key:        conf.Key,
		clientID:   conf.ClientID,
	}
}

func (c *Client) Run() {
	for {
		err := c.run()
		if err != nil && err != io.EOF {
			logrus.Errorf("run error: %v", err)
		}
		logrus.Warnf("Reconnecting...")
		time.Sleep(time.Second * 2)
	}
}

func (c *Client) run() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	handshakeReq := common.HandshakeReq{ClientID: c.clientID}
	buf, err := handshakeReq.Encode()
	if err != nil {
		return err
	}
	// 发送 handshake 请求
	conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	_, err = conn.Write(buf)
	conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}

	// 创建 smux
	smuxconfig := smux.DefaultConfig()
	smuxconfig.KeepAliveInterval = 1 * time.Second
	smuxconfig.KeepAliveTimeout = 2 * time.Second
	mux, err := smux.Client(conn, smuxconfig)
	if err != nil {
		return err
	}
	logrus.Debug("Handshake success！")
	logrus.Debugf("Success connect server: %s", c.serverAddr)
	defer mux.Close()
	for {
		stream, err := mux.AcceptStream()
		if err != nil {
			return err
		}
		go c.handleStream(stream)
	}
}

func (c *Client) handleStream(tunnelConn common.VeilConn) {
	defer tunnelConn.Close()
	enc := &common.EncryptProtocl{}
	encryptOn, err := enc.Check(tunnelConn)
	if err != nil {
		logrus.Errorf("Check error: %v", err)
		return
	}
	if encryptOn {
		byteKey, err := common.KeyStringToByte(c.key)
		if err != nil {
			logrus.Errorf("KeyStringToByte error: %v", err)
			return
		}
		tunnelConn, err = common.NewChacha20Stream(byteKey, tunnelConn)
		if err != nil {
			logrus.Errorf("NewChacha20Stream error: %v", err)
			return
		}
	}

	vp := &common.VeilinkProtocol{}
	if err = vp.Decode(tunnelConn); err != nil {
		logrus.Errorf("Decode error: %v", err)
		return
	}

	var localConn net.Conn
	switch vp.PublicProtocol {
	case "tcp":
		localConn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", vp.InternalIP, vp.InternalPort))
		if err != nil {
			logrus.Errorf("Dial error: %v", err)
			return
		}
		defer localConn.Close()

		in, out := common.Join(localConn, tunnelConn)
		logrus.Infof("in: %d bytes, out: %d bytes", in, out)
	case "udp":
		localConn, err = net.Dial("udp", fmt.Sprintf("%s:%d", vp.InternalIP, vp.InternalPort))
		if err != nil {
			logrus.Errorf("Dial error: %v", err)
			return
		}
		go func() {
			defer localConn.Close()
			defer tunnelConn.Close()
			buf := make([]byte, 1024*64)
			for {
				nr, err := localConn.Read(buf)
				if err != nil {
					logrus.Errorf("Decode error: %v", err)
					break
				}
				p := common.UDPpacket(buf[:nr])
				body, err := p.Encode()
				if err != nil {
					logrus.Errorf("Encode error: %v", err)
					break
				}

				_, err = tunnelConn.Write(body)
				if err != nil {
					logrus.Errorf("Write error: %v", err)
					break
				}
			}
		}()

		p := common.UDPpacket{}
		for {
			err := p.Decode(tunnelConn)
			if err != nil {
				logrus.Errorf("Decode error: %v", err)
				break
			}
			_, err = localConn.Write(p)
			if err != nil {
				logrus.Errorf("Write error: %v", err)
				break
			}
		}
	default:
		logrus.Warnf("Unsupported protocol: %s", vp.PublicProtocol)
		return
	}

}
