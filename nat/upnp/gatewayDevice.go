package upnp

import (
	"strings"
	"time"
)

// Hack to avoid Go complaining if time isn't used.
var _ time.Time

// Device URNs:
const (
	URN_DEVICE_LANDevice_1           = "urn:schemas-upnp-org:device:LANDevice:1"
	URN_DEVICE_WANConnectionDevice_1 = "urn:schemas-upnp-org:device:WANConnectionDevice:1"
	URN_DEVICE_WANConnectionDevice_2 = "urn:schemas-upnp-org:device:WANConnectionDevice:2"
	URN_DEVICE_WANDevice_1           = "urn:schemas-upnp-org:device:WANDevice:1"
	URN_DEVICE_WANDevice_2           = "urn:schemas-upnp-org:device:WANDevice:2"
)

// Service URNs:
const (
	URN_SERVICE_DeviceProtection_1         = "urn:schemas-upnp-org:service:DeviceProtection:1"
	URN_SERVICE_LANHostConfigManagement_1  = "urn:schemas-upnp-org:service:LANHostConfigManagement:1"
	URN_SERVICE_Layer3Forwarding_1         = "urn:schemas-upnp-org:service:Layer3Forwarding:1"
	URN_SERVICE_WANCableLinkConfig_1       = "urn:schemas-upnp-org:service:WANCableLinkConfig:1"
	URN_SERVICE_WANCommonInterfaceConfig_1 = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1"
	URN_SERVICE_WANDSLLinkConfig_1         = "urn:schemas-upnp-org:service:WANDSLLinkConfig:1"
	URN_SERVICE_WANEthernetLinkConfig_1    = "urn:schemas-upnp-org:service:WANEthernetLinkConfig:1"
	URN_SERVICE_WANIPv6FirewallControl_1   = "urn:schemas-upnp-org:service:WANIPv6FirewallControl:1"
	URN_SERVICE_WANPOTSLinkConfig_1        = "urn:schemas-upnp-org:service:WANPOTSLinkConfig:1"
	URN_SERVICE_WANIPConnection_1          = "urn:schemas-upnp-org:service:WANIPConnection:1"
	URN_SERVICE_WANIPConnection_2          = "urn:schemas-upnp-org:service:WANIPConnection:2"
	URN_SERVICE_WANPPPConnection_1         = "urn:schemas-upnp-org:service:WANPPPConnection:1"
)

func IsInternalGatewayDevice(st string) bool {
	return strings.Contains(st, "InternetGatewayDevice")
}
