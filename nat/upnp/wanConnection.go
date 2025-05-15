package upnp

import (
	"i4remoter/pkg/nat/soap"
)

type WANConnection struct {
	soapClient *soap.SOAPClient
	serverType string
}

func NewWANConnection(soapClient *soap.SOAPClient, serverType string) *WANConnection {
	return &WANConnection{
		soapClient: soapClient,
		serverType: serverType,
	}
}

func (sel *WANConnection) getServiceType() string {
	return sel.serverType
}

func (sel *WANConnection) AddPortMapping(req *AddPortMappingReq) (err error) {
	return sel.soapClient.DoCall(sel.getServiceType(), ActionAddPortMapping, req, nil)
}

func (sel *WANConnection) DeletePortMapping(remoteHost string, externalPort uint16, protocol string) (err error) {
	// Request structure.
	request := &DeletePortMappingReq{
		NewRemoteHost:   remoteHost,
		NewProtocol:     protocol,
		NewExternalPort: externalPort,
	}
	response := interface{}(nil)
	return sel.soapClient.DoCall(sel.getServiceType(), ActionDeletePortMapping, request, response)
}

func (sel *WANConnection) GetExternalIPAddress() (externalIPAddress string, err error) {
	// Request structure.
	request := struct{}{}
	// Response structure.
	response := &struct {
		NewExternalIPAddress string
	}{}
	// Perform the SOAP call.
	if err = sel.soapClient.DoCall(sel.getServiceType(), ActionGetExternalIPAddress, request, response); err != nil {
		return
	}
	externalIPAddress = response.NewExternalIPAddress
	return
}

func (sel *WANConnection) GetConnectionTypeInfo() (*GetConnectionTypeInfoResp, error) {
	// Request structure.
	request := struct{}{}
	// Response structure.
	response := &GetConnectionTypeInfoResp{}
	// Perform the SOAP call.
	if err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetConnectionTypeInfo, request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (sel *WANConnection) GetGenericPortMappingEntry(mappingIndex uint16) (*GetGenericPortMappingEntryResp, error) {
	// Request structure.
	request := &GetGenericPortMappingEntryReq{
		NewPortMappingIndex: mappingIndex,
	}

	// Response structure.
	response := &GetGenericPortMappingEntryResp{}
	// Perform the SOAP call.
	if err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetGenericPortMappingEntry, request, response); err != nil {
		return nil, err
	}

	return response, nil
}

func (sel *WANConnection) GetSpecificPortMappingEntry(remoteHost string, externalPort uint16, protocol string) (*GetSpecificPortMappingEntryResp, error) {
	// Request structure.
	request := &GetSpecificPortMappingEntryReq{
		NewRemoteHost:   remoteHost,
		NewExternalPort: externalPort,
		NewProtocol:     protocol,
	}
	// Response structure.
	response := &GetSpecificPortMappingEntryResp{}
	// Perform the SOAP call.
	err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetSpecificPortMappingEntry, request, response)

	return response, err
}

func (sel *WANConnection) GetStatusInfo() (*GetStatusInfoResp, error) {
	// Request structure.
	request := &GetStatusInfoReq{}
	// Response structure.
	response := &GetStatusInfoResp{}
	err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetStatusInfo, request, response)
	return response, err
}
