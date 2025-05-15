package ssdp

import (
	"net"
)

func GetInterfacesIPv4() ([]*net.Interface, error) {
	itfList, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	list := make([]*net.Interface, 0, len(itfList))
	for _, ifi := range itfList {
		if !hasLinkUp(&ifi) || !hasMulticast(&ifi) || !hasIPv4Address(&ifi) || !hasUp(&ifi) {
			continue
		}
		list = append(list, &ifi)
	}
	return list, nil
}

func hasLinkUp(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagUp != 0
}

func hasMulticast(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagMulticast != 0
}

// 启用状态
func hasUp(ifi *net.Interface) bool {
	return ifi.Flags&net.FlagUp != 0
}

func hasIPv4Address(ifi *net.Interface) bool {
	addrs, err := ifi.Addrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			continue
		}
		if len(ip.To4()) == net.IPv4len && !ip.IsUnspecified() {
			return true
		}
	}
	return false
}
