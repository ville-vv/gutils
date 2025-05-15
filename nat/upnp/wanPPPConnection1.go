package upnp

import (
	"i4remoter/pkg/nat/soap"
)

type WANPPPConnection1 struct {
	*WANConnection
}

func NewWANPPPConnection1(soapClient *soap.SOAPClient) *WANPPPConnection1 {
	return &WANPPPConnection1{
		WANConnection: NewWANConnection(soapClient, URN_SERVICE_WANPPPConnection_1),
	}
}

func (sel *WANPPPConnection1) GetPPPCompressionProtocol() (string, error) {
	// Request structure.
	request := interface{}(nil)
	// Response structure.
	response := &GetPPPCompressionProtocolResp{}

	// Perform the SOAP call.
	if err := sel.soapClient.DoCall(sel.getServiceType(), ActionGetPPPCompressionProtocol, request, response); err != nil {
		return "", err
	}

	// END Unmarshal arguments from response.
	return response.NewPPPCompressionProtocol, nil
}
