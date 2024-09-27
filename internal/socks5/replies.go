package socks5

import (
	"io"
	"net"
)

type REP = byte

const (
	REP_SUCCEEDED                         REP = 0x00 // 成功
	REP_GENERAL_SOCKS_SERVER_FAILURE      REP = 0x01 // 常规 SOCKS 服务故障
	REP_CONNECTION_NOT_ALLOWED_BY_RULESET REP = 0x02 // 规则集不允许连接
	REP_NETWORK_UNREACHABLE               REP = 0x03 // 网络不可达
	REP_HOST_UNREACHABLE                  REP = 0x04 // 主机不可达
	REP_CONNECTION_REFUSED                REP = 0x05 // 连接被拒绝
	REP_TTL_EXPIRED                       REP = 0x06 // 连接超时
	REP_COMMAND_NOT_SUPPORTED             REP = 0x07 // 命令不支持
	REP_ADDRESS_TYPE_NOT_SUPPORTED        REP = 0x08 // 地址类型不支持
	REP_UNASSIGNED                        REP = 0x09 // 未定义
)

/*
		 	+----+-----+-------+------+----------+----------+
	        |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	        +----+-----+-------+------+----------+----------+
	        | 1  |  1  | X'00' |  1   | Variable |    2     |
	        +----+-----+-------+------+----------+----------+
*/
type replies []byte

func RepliesTODO() replies {
	return []byte{socks5Version, REP_SUCCEEDED, RSV, IPv4, 0, 0, 0, 0, 0, 0}
}

func (r replies) WithREPByError(err error) replies {
	switch err {
	case UnSupportCMD:
		return r.WithREP(REP_COMMAND_NOT_SUPPORTED)
	case UnSupportATYP:
		return r.WithREP(REP_ADDRESS_TYPE_NOT_SUPPORTED)
	default:
		return r.WithREP(REP_GENERAL_SOCKS_SERVER_FAILURE)
	}
}

func (r replies) WithREP(rep REP) replies {
	r[1] = rep
	return r
}

func (r replies) WithATYP(atyp ATYP) replies {
	r[3] = atyp
	return r
}

func (r replies) WithADDR(netaddr net.Addr) replies {
	// TODO
	return r
}

func (r replies) Wirte(conn io.Writer) error {
	_, err := conn.Write(r)
	return err
}
