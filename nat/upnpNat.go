package nat

import (
	upnp "i4remoter/pkg/nat/upnp"
	"net"
	"net/url"
	"time"
)

var _ NAT = (*upnpNat)(nil)

type upnpNat struct {
	upnpCli    upnpNATClient
	deviceName string
	natType    string
	location   url.URL
	rootDevice *upnp.RootDevice
}

func (sel *upnpNat) Type() string {
	return sel.natType
}

func (sel *upnpNat) Location() *url.URL {
	return &sel.location
}

func (sel *upnpNat) DeviceName() string {
	return sel.deviceName
}

func (sel *upnpNat) GetDeviceAddress() (net.IP, error) {
	addr, err := net.ResolveUDPAddr("udp4", sel.location.Host)
	if err != nil {
		return nil, err
	}
	return addr.IP, nil
}

func (sel *upnpNat) GetExternalAddress() (addr net.IP, err error) {
	ipString, err := sel.upnpCli.GetExternalIPAddress()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(ipString)
	if ip == nil {
		return nil, ErrNoExternalAddress
	}
	return ip, nil
}

func (sel *upnpNat) GetInternalAddress() (addr net.IP, err error) {
	devAddr, err := sel.GetDeviceAddress()
	if err != nil {
		return nil, err
	}
	return InGatewayLocalIp(devAddr)
}

func (sel *upnpNat) AddPortMapping(protocol string, internalPort int, externalPort int, timeout uint, desc string) (err error) {
	//if err = isSupportProtocol(Protocol(protocol)); err != nil {
	//	return err
	//}
	ip, err := sel.GetInternalAddress()
	if err != nil {
		return nil
	}
	if desc == "" {
		desc = portMappingDescription
	}
	req := &upnp.AddPortMappingReq{
		NewProtocol:               Protocol(protocol).String(),
		NewExternalPort:           uint16(externalPort),
		NewInternalPort:           uint16(internalPort),
		NewInternalClient:         ip.String(),
		NewEnabled:                1,
		NewPortMappingDescription: desc,
		NewLeaseDuration:          uint32(timeout),
	}
	for i := 0; i < 3; i++ {
		if err = sel.upnpCli.AddPortMapping(req); err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 50)
	}
	return err
}

func (sel *upnpNat) DeletePortMapping(protocol string, externalPort int) (err error) {
	if err = Protocol(protocol).IsSupport(); err != nil {
		return err
	}
	return sel.upnpCli.DeletePortMapping("", uint16(externalPort), Protocol(protocol).String())
}

func (sel *upnpNat) GetPortMapping(protocol string, externalPort int) (*upnp.GetSpecificPortMappingEntryResp, error) {
	return sel.upnpCli.GetSpecificPortMappingEntry("", uint16(externalPort), Protocol(protocol).String())
}

func (sel *upnpNat) GetDeviceStatus() (*upnp.GetStatusInfoResp, error) {
	return sel.upnpCli.GetStatusInfo()
}
