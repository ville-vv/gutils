package upnp

import "testing"

func TestWANPPPConnection1_AddPortMapping(t *testing.T) {
	tests := []struct {
		name string
		args *AddPortMappingReq
	}{
		{
			name: "",
			args: &AddPortMappingReq{
				NewRemoteHost:             "",
				NewExternalPort:           47890,
				NewProtocol:               "TCP",
				NewInternalPort:           47890,
				NewInternalClient:         "",
				NewEnabled:                1,
				NewPortMappingDescription: "",
				NewLeaseDuration:          60,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WANPPPConnection1{}
			if err := w.AddPortMapping(tt.args); err != nil {
				t.Errorf("WANPPPConnection1.AddPortMapping() error = %v", err)
			}
		})
	}
}
