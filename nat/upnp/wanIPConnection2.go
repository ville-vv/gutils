package upnp

import (
	"i4remoter/pkg/nat/soap"
)

type WANIPConnection2 struct {
	*WANIPConnection1
}

func NewWANIPConnection2(soapClient *soap.SOAPClient) *WANIPConnection2 {
	return &WANIPConnection2{
		WANIPConnection1: &WANIPConnection1{
			WANConnection: NewWANConnection(soapClient, URN_SERVICE_WANIPConnection_2),
		},
	}
}

func (sel *WANIPConnection2) AddAnyPortMapping(request *AddAnyPortMappingReq) (uint16, error) {
	response := &struct {
		NewReservedPort uint16
	}{}
	err := sel.soapClient.DoCall(sel.getServiceType(), ActionAddAnyPortMapping, request, response)
	// END Unmarshal arguments from response.
	return response.NewReservedPort, err
}

// DeletePortMappingRange
// Arguments: * NewProtocol: allowed values: TCP, UDP
func (sel *WANIPConnection2) DeletePortMappingRange(startPort uint16, endPort uint16, protocol string, manage bool) (err error) {
	request := &DeletePortMappingRangeReq{
		NewStartPort: startPort,
		NewEndPort:   endPort,
		NewProtocol:  protocol,
	}
	if manage {
		request.NewManage = 1
	}
	if err = sel.soapClient.DoCall(sel.getServiceType(), ActionDeletePortMappingRange, request, nil); err != nil {
		return
	}
	return
}

func (sel *WANIPConnection2) GetListOfPortMappings(request *GetListOfPortMappingsReq) (portListing string, err error) {
	response := &GetListOfPortMappingsResp{}
	if err = sel.soapClient.DoCall(sel.getServiceType(), ActionGetListOfPortMappings, request, response); err != nil {
		return
	}
	portListing = response.NewPortListing
	return
}
