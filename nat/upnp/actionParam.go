package upnp

type AddPortMappingReq struct {
	NewRemoteHost             string `action:"AddPortMapping"`
	NewProtocol               string
	NewExternalPort           uint16
	NewInternalPort           uint16
	NewInternalClient         string
	NewEnabled                int
	NewPortMappingDescription string
	NewLeaseDuration          uint32
}

type AddAnyPortMappingReq struct {
	NewRemoteHost             string `action:"AddAnyPortMapping"`
	NewProtocol               string
	NewExternalPort           uint16
	NewInternalPort           uint16
	NewInternalClient         string
	NewEnabled                int
	NewPortMappingDescription string
	NewLeaseDuration          uint32
}

type DeletePortMappingReq struct {
	NewRemoteHost   string `action:"DeletePortMapping"`
	NewProtocol     string
	NewExternalPort uint16
}

type DeletePortMappingRangeReq struct {
	NewStartPort uint16
	NewEndPort   uint16
	NewProtocol  string
	NewManage    uint
}

type GetGenericPortMappingEntryReq struct {
	NewPortMappingIndex uint16 `action:"GetGenericPortMappingEntry"`
}

type GetGenericPortMappingEntryResp struct {
	NewRemoteHost             string
	NewProtocol               string
	NewExternalPort           uint16
	NewInternalPort           uint16
	NewInternalClient         string
	NewEnabled                int
	NewPortMappingDescription string
	NewLeaseDuration          int
}

type GetStatusInfoReq struct {
	soapAction string `action:"GetStatusInfo"`
}
type GetStatusInfoResp struct {
	NewConnectionStatus    string
	NewLastConnectionError string
	NewUptime              uint32
}

type GetSpecificPortMappingEntryReq struct {
	NewRemoteHost   string `action:"GetSpecificPortMappingEntry"`
	NewExternalPort uint16
	NewProtocol     string
}

type GetSpecificPortMappingEntryResp struct {
	NewInternalPort           uint16
	NewInternalClient         string
	NewEnabled                int
	NewPortMappingDescription string
	NewLeaseDuration          uint32
}

type GetConnectionTypeInfoResp struct {
	NewConnectionType          string
	NewPossibleConnectionTypes string
}

type GetPPPCompressionProtocolResp struct {
	NewPPPCompressionProtocol string
}

type GetListOfPortMappingsReq struct {
	NewStartPort     uint16
	NewEndPort       uint16
	NewProtocol      string
	NewManage        int
	NewNumberOfPorts uint16
}
type GetListOfPortMappingsResp struct {
	NewPortListing string
}

type GetNATRSIPStatusResp struct {
	NewRSIPAvailable string
	NewNATEnabled    string
}
