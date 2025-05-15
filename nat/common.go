package nat

import (
	"errors"
	"i4remoter/pkg/nat/upnp"
	"net"
	"strings"
)

const (
	ProtocolTCP Protocol = "TCP" // ProtocolTCP tcp协议
	ProtocolUDP Protocol = "UDP" // ProtocolUDP udp协议
)

var (
	portMappingDescription = "i4remote" // 使用 var 定义是方便外部修改
)

var (
	ErrNotSupportProtocol = errors.New("not support protocol")
	ErrNoExternalAddress  = errors.New("no external address")
	ErrNoInternalAddress  = errors.New("no internal address")
)

type Protocol string

func (p Protocol) String() string {
	return strings.ToUpper(string(p))
}

func (p Protocol) IsSupport() error {
	switch Protocol(p.String()) {
	case ProtocolTCP, ProtocolUDP:
		return nil
	default:
		return ErrNotSupportProtocol
	}
}

func InGatewayLocalIp(gatewayIp net.IP) (net.IP, error) {
	itfs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, itf := range itfs {
		addrs, err := itf.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch x := addr.(type) {
			case *net.IPNet:
				if x.Contains(gatewayIp) {
					return x.IP, nil
				}
			}
		}
	}
	return nil, ErrNoInternalAddress
}

type upnpNATClient interface {
	GetExternalIPAddress() (externalIPAddress string, err error)
	AddPortMapping(req *upnp.AddPortMappingReq) error
	DeletePortMapping(remoteHost string, externalPort uint16, protocol string) error
	GetSpecificPortMappingEntry(remoteHost string, externalPort uint16, protocol string) (*upnp.GetSpecificPortMappingEntryResp, error)
	GetStatusInfo() (*upnp.GetStatusInfoResp, error)
}
