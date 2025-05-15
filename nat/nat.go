package nat

import (
	"i4remoter/pkg/nat/upnp"
	"net"
	"net/url"
)

// NAT protocol is either "udp" or "tcp"
type NAT interface {
	// Type 什么类型的NAT设备
	Type() string

	DeviceName() string

	Location() *url.URL

	// GetDeviceAddress 获取NAT设备的话网关地址,设备地址
	GetDeviceAddress() (addr net.IP, err error)

	// GetExternalAddress 获取NAT设备的外网地址
	GetExternalAddress() (addr net.IP, err error)

	// GetInternalAddress 获取当前连接NAT设备的内网地址
	GetInternalAddress() (addr net.IP, err error)

	// AddPortMapping 添加端口映射
	AddPortMapping(protocol string, internalPort int, externalPort int, leaseDuration uint, desc string) (err error)

	// DeletePortMapping 移除端口映射
	DeletePortMapping(protocol string, internalPort int) (err error)

	GetPortMapping(protocol string, externalPort int) (*upnp.GetSpecificPortMappingEntryResp, error)

	GetDeviceStatus() (*upnp.GetStatusInfoResp, error)
}
