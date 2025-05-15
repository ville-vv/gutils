package soap

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncodeAction(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	action := &actionBuilder{
		//ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:1",
		//ActionName:  "AddPortMapping",
		Space: "urn:schemas-upnp-org:service:WANIPConnection:1",
		//Args: &PortMappingRequest{
		//	ExternalPort:   4651,
		//	InternalPort:   4651,
		//	Protocol:       "tcp",
		//	Enabled:        1,
		//	InternalClient: "123.0.0.1",
		//	LeaseDuration:  0,
		//	Description:    "mandela",
		//	RemoteHost:     "", // 明确保留空字段
		//},
		Args: map[string]interface{}{
			"NewExternalPort": 4651,
			"NewInternalPort": 4651,
			"NewProtocol":     "tcp",
			"NewEnabled":      1,
			"InternalClient":  "123.0.0.1",
		},
	}
	EncodeAction(buf, action)

	fmt.Println(string(buf.Bytes()))

	actionOut := &actionBuilder{
		//Args: &PortMappingRequest{},
	}
	DecodeAction(buf, actionOut)
	fmt.Println(actionOut.Args)

}
