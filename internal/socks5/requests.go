package socks5

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"slices"
)

var (
	RSVError    = errors.New("read rsv error")
	UnknownCMD  = errors.New("unknown command")
	UnknownATYP = errors.New("unknown auth method")
)

type CMD = byte

const (
	Connect CMD = 0x01 // 
	Bind    CMD = 0x02
	UdpAss  CMD = 0x03
)

var cmds = []byte{Connect, Bind, UdpAss}

type ATYP = byte

const (
	IPv4   ATYP = 0x01
	Domain ATYP = 0x03
	IPv6   ATYP = 0x04
)

var atyps = []byte{IPv4, Domain, IPv6}

const RSV byte = 0x00

/*
		+----+-----+-------+------+----------+----------+
	    |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	    +----+-----+-------+------+----------+----------+
	    | 1  |  1  | X'00' |  1   | Variable |    2     |
	    +----+-----+-------+------+----------+----------+
*/
type ClientRequestMessage struct {
	Version byte   // socks5版本
	Cmd     CMD    // 命令
	Rsv     byte   // 保留
	Atyp    ATYP   // 地址类型
	DstAddr string // 目标地址
	DstPort int    // 目标端口
}

func NewClientRequestMessage(conn io.Reader) (*ClientRequestMessage, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	// 验证
	ver, cmd, rsv, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != socks5Version {
		return nil, UnsupportedProtocol
	}

	if !slices.Contains(cmds, cmd) {
		return nil, UnknownCMD
	}

	if rsv != RSV {
		return nil, RSVError
	}

	if !slices.Contains(atyps, atyp) {
		return nil, UnknownATYP
	}

	// 读取IP地址
	var addrbyte []byte
	var err error
	switch atyp {
	case IPv4:
		addrbyte = make([]byte, net.IPv4len)
		_, err = io.ReadFull(conn, addrbyte)
		if err != nil {
			return nil, err
		}
	case IPv6:
		addrbyte = make([]byte, net.IPv6len)
		_, err = io.ReadFull(conn, addrbyte)
		if err != nil {
			return nil, err
		}
	case Domain:
		// 读取域名长度
		addrlenbuf := make([]byte, 1)
		if _, err = io.ReadFull(conn, addrlenbuf); err != nil {
			return nil, err
		}
		// 读取域名
		addrbuf := make([]byte, addrlenbuf[0])
		if _, err = io.ReadFull(conn, addrbuf); err != nil {
			return nil, err
		}
		// 解析域名
		netaddr, err := net.ResolveIPAddr("ip", string(addrbuf))
		if err != nil {
			return nil, err
		}
		addrbyte = netaddr.IP
	}

	// 读取端口
	portbuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portbuf); err != nil {
		return nil, err
	}
	port := int(binary.BigEndian.Uint16(portbuf))

	return &ClientRequestMessage{
		Version: ver,
		Cmd:     cmd,
		Rsv:     rsv,
		Atyp:    atyp,
		DstAddr: net.IP(addrbyte).String(),
		DstPort: port,
	}, nil
}
