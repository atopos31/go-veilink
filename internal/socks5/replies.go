package socks5

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

type replies []byte

func RepliesTODO() replies {
	return []byte{socks5Version, REP_SUCCEEDED, RSV, IPv4}
}

func (r replies) WithREP(rep REP) replies {
	r[1] = rep
	return r
}

func (r replies) WithATYP(atyp ATYP) replies {
	r[3] = atyp
	return r
}
