package upnp

import (
	"i4remoter/pkg/nat/soap"
)

type WANIPConnection1 struct {
	*WANConnection
}

func NewWANIPConnection1(soapClient *soap.SOAPClient) *WANIPConnection1 {
	return &WANIPConnection1{
		WANConnection: NewWANConnection(soapClient, URN_SERVICE_WANIPConnection_1),
	}
}

func (sel *WANIPConnection1) ForceTermination() (err error) {
	return sel.soapClient.DoCall(sel.getServiceType(), ActionForceTermination, nil, nil)
}

func (sel *WANIPConnection1) GetIdleDisconnectTime() (idleDisconnectTime uint32, err error) {
	response := &struct {
		NewIdleDisconnectTime uint32
	}{}
	if err = sel.soapClient.DoCall(sel.getServiceType(), "GetIdleDisconnectTime", nil, response); err != nil {
		return
	}
	idleDisconnectTime = response.NewIdleDisconnectTime
	return
}

func (sel *WANIPConnection1) GetNATRSIPStatus() (*GetNATRSIPStatusResp, error) {
	response := &GetNATRSIPStatusResp{}
	err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetNATRSIPStatus, nil, response)
	return response, err
}

func (sel *WANIPConnection1) GetWarnDisconnectDelay() (NewWarnDisconnectDelay uint32, err error) {
	request := interface{}(nil)
	response := &struct {
		NewWarnDisconnectDelay uint32
	}{}

	if err = sel.soapClient.DoCall(sel.getServiceType(), "GetWarnDisconnectDelay", request, response); err != nil {
		return
	}
	NewWarnDisconnectDelay = response.NewWarnDisconnectDelay
	return
}

func (sel *WANIPConnection1) RequestConnection() (err error) {
	return sel.soapClient.DoCall(sel.getServiceType(), "RequestConnection", nil, nil)
}

func (sel *WANIPConnection1) RequestTermination() (err error) {
	return sel.soapClient.DoCall(sel.getServiceType(), "RequestTermination", nil, nil)
}

func (sel *WANIPConnection1) SetAutoDisconnectTime(NewAutoDisconnectTime uint32) (err error) {
	request := &struct {
		NewAutoDisconnectTime uint32
	}{
		NewAutoDisconnectTime: NewAutoDisconnectTime,
	}

	return sel.soapClient.DoCall(sel.getServiceType(), "SetAutoDisconnectTime", request, nil)
}

func (sel *WANIPConnection1) SetConnectionType(NewConnectionType string) (err error) {
	request := &struct {
		NewConnectionType string
	}{
		NewConnectionType: NewConnectionType,
	}
	return sel.soapClient.DoCall(sel.getServiceType(), "SetConnectionType", request, nil)
}

func (sel *WANIPConnection1) SetIdleDisconnectTime(NewIdleDisconnectTime uint32) (err error) {
	request := &struct {
		NewIdleDisconnectTime uint32
	}{
		NewIdleDisconnectTime: NewIdleDisconnectTime,
	}
	return sel.soapClient.DoCall(sel.getServiceType(), "SetIdleDisconnectTime", request, nil)
}

func (sel *WANIPConnection1) SetWarnDisconnectDelay(NewWarnDisconnectDelay uint32) (err error) {
	request := &struct {
		NewWarnDisconnectDelay uint32
	}{
		NewWarnDisconnectDelay: NewWarnDisconnectDelay,
	}
	return sel.soapClient.DoCall(sel.getServiceType(), "SetWarnDisconnectDelay", request, nil)
}
