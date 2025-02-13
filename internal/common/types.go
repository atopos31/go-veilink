package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

const (
	version      = 0
	cmdVP        = 0x0
	cmdHandshake = 0x1
	cmdHandudp   = 0x2
)

const (
	cmdEncrypt   = 0x4
	cmdEncryptOn = 0x1
	cmdEncryptOf = 0x0
)

var (
	ErrVersion   = errors.New("Invalid vp version error")
	ErrCmd       = errors.New("Invalid vp cmd error")
	ErrHandshake = errors.New("Invalid vp handshake error")
	ErrHandudp   = errors.New("Invalid vp Handudp error")

	ErrEncrypt = errors.New("Invalid vp Encrypt error")
)

// VeilinkProtocol Veilink协议
type VeilinkProtocol struct {
	ClientID       string // 客户端ID
	PublicProtocol string // 外网协议
	PublicIP       string // 外网IP
	PublicPort     uint16 // 外网端口
	InternalIP     string // 内网IP
	InternalPort   uint16 // 内网端口
}

func (vp *VeilinkProtocol) Encode() ([]byte, error) {
	header := make([]byte, 4)
	header[0] = version
	header[1] = cmdVP

	body, err := json.Marshal(vp)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(header[2:4], uint16(len(body)))
	return append(header, body...), nil
}

func (vp *VeilinkProtocol) Decode(reader io.Reader) error {
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return err
	}
	if header[0] != version {
		return ErrVersion
	}
	if header[1] != cmdVP {
		return ErrCmd
	}

	bodyLen := binary.BigEndian.Uint16(header[2:4])
	body := make([]byte, bodyLen)
	if _, err := io.ReadFull(reader, body); err != nil {
		return err
	}
	return json.Unmarshal(body, vp)
}

type HandshakeReq struct {
	ClientID string
}

func (req *HandshakeReq) Encode() ([]byte, error) {
	hdr := make([]byte, 4)
	hdr[0] = version
	hdr[1] = cmdHandshake

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(hdr[2:4], uint16(len(body)))
	return append(hdr, body...), nil
}

func (req *HandshakeReq) Decode(reader io.Reader) error {
	hdr := make([]byte, 4)
	_, err := io.ReadFull(reader, hdr)
	if err != nil {
		return err
	}

	cmd := hdr[1]
	if cmd != cmdHandshake {
		return ErrHandshake
	}

	bodyLen := binary.BigEndian.Uint16(hdr[2:4])

	body := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		return err
	}

	return nil
}

type UDPpacket []byte

func (pkt UDPpacket) Encode() ([]byte, error) {
	hdr := make([]byte, 4)
	hdr[0] = version
	hdr[1] = cmdHandudp

	binary.BigEndian.PutUint16(hdr[2:4], uint16(len(pkt)))
	return append(hdr, pkt...), nil
}

func (pkt *UDPpacket) Decode(reader io.Reader) error {
	hdr := make([]byte, 4)
	hdr[0] = version
	hdr[1] = cmdHandudp

	_, err := io.ReadFull(reader, hdr)
	if err != nil {
		return err
	}

	cmd := hdr[1]
	if cmd != cmdHandudp {
		return ErrHandudp
	}

	bodyLen := binary.BigEndian.Uint16(hdr[2:4])
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return err
	}
	*pkt = body

	return nil
}

type EncryptProtocl []byte

func (ep EncryptProtocl) Encode(envOn bool) []byte {
	hdr := make([]byte, 2)
	hdr[0] = cmdEncrypt
	if envOn {
		hdr[1] = cmdEncryptOn
	} else {
		hdr[1] = cmdEncryptOf
	}
	return hdr
}

func (ep EncryptProtocl) Check(reader io.Reader) (bool, error) {
	hdr := make([]byte, 2)
	_, err := io.ReadFull(reader, hdr)
	if err != nil {
		return false, err
	}

	cmd := hdr[0]
	if cmd != cmdEncrypt {
		return false, ErrEncrypt
	}

	encrypt := hdr[1]
	if encrypt == cmdEncryptOn {
		return true, nil
	} else if encrypt == cmdEncryptOf {
		return false, nil
	} else {
		return false, ErrEncrypt
	}
}
