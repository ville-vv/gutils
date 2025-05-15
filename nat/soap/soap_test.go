package soap

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

type capturingRoundTripper struct {
	err         error
	resp        *http.Response
	capturedReq *http.Request
}

func (rt *capturingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.capturedReq = req
	return rt.resp, rt.err
}

func TestActionInputs(t *testing.T) {
	t.Parallel()
	url, err := url.Parse("http://example.com/soap")
	if err != nil {
		t.Fatal(err)
	}
	respBodyBuf := bytes.NewBufferString(`
				<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
					<s:Body>
						<u:myactionResponse xmlns:u="mynamespace">
							<A>valueA</A>
							<B>valueB</B>
						</u:myactionResponse>
					</s:Body>
				</s:Envelope>
			`)
	rt := &capturingRoundTripper{
		err: nil,
		resp: &http.Response{
			StatusCode:    200,
			Body:          io.NopCloser(respBodyBuf),
			ContentLength: int64(respBodyBuf.Len()),
		},
	}

	client := SOAPClient{
		controlURL: *url,
		httpClient: http.Client{
			Transport: rt,
		},
	}

	type In struct {
		Foo string
		Bar string `soap:"bar"`
		Baz string `xml:"baz"`
	}
	type Out struct {
		A string
		B string
	}

	in := In{"foo", "bar", "quoted=\"baz\""}
	gotOut := Out{}
	err = client.DoCall("mynamespace", "myaction", &in, &gotOut)
	if err != nil {
		t.Fatal(err)
	}

	wantBody := (soapPrefix +
		`<u:myaction xmlns:u="mynamespace">` +
		`<Foo>foo</Foo>` +
		`<Bar>bar</Bar>` +
		`<baz>quoted="baz"</baz>` +
		`</u:myaction>` +
		soapSuffix)
	body, err := io.ReadAll(rt.capturedReq.Body)
	if err != nil {
		t.Fatal(err)
	}
	gotBody := string(body)
	if wantBody != gotBody {
		t.Errorf("Bad request body\nwant: %q\n got: %q", wantBody, gotBody)
	}

	wantOut := Out{"valueA", "valueB"}
	if !reflect.DeepEqual(wantOut, gotOut) {
		t.Errorf("Bad output\nwant: %+v\n got: %+v", wantOut, gotOut)
	}
}

func TestDecodeToAction_OutError(t *testing.T) {
	buf := bytes.NewBufferString(`<s:Envelope
xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"
s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<s:Fault>
<faultcode>s:Client</faultcode>
<faultstring>UPnPError</faultstring>
<detail>
<UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
<errorCode>713</errorCode>
<errorDescription>SpecifiedArrayIndexInvalid</errorDescription>
</UPnPError>
</detail>
</s:Fault>
</s:Body>
</s:Envelope>
`)
	fmt.Println(DecodeAction(buf, newRecvAction(nil)))
}

func TestDecodeToAction_OutError2(t *testing.T) {
	buf := bytes.NewBufferString(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:GetGenericPortMappingEntryResponse xmlns:u="urn:schemas-upnp-org:service:WANPPPConnection:1">
<NewRemoteHost/>
<NewExternalPort>28571</NewExternalPort>
<NewProtocol>UDP</NewProtocol>
<NewInternalPort>28571</NewInternalPort>
<NewInternalClient>192.168.3.7</NewInternalClient>
<NewEnabled>1</NewEnabled>
<NewPortMappingDescription>192.168.3.7@28571</NewPortMappingDescription>
<NewLeaseDuration>604558</NewLeaseDuration>
</u:GetGenericPortMappingEntryResponse>
</s:Body>
</s:Envelope>
`)
	response := struct {
		NewRemoteHost             string
		NewExternalPort           uint16
		NewProtocol               string
		NewInternalPort           uint16
		NewInternalClient         string
		NewEnabled                bool
		NewPortMappingDescription string
		NewLeaseDuration          uint32
	}{}
	fmt.Println(DecodeAction(buf, newRecvAction(&response)))
	fmt.Println("Result:", response)
}

func TestDecodeToAction_OutOk(t *testing.T) {
	buf := bytes.NewBufferString(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>
<u:AddPortMappingResponse xmlns:u="urn:schemas-upnp-org:service:WANPPPConnection:1"/>
</s:Body>
</s:Envelope>
`)
	fmt.Println(DecodeAction(buf, newRecvAction(nil)))
}

func TestDecodeToAction_Empty(t *testing.T) {
	buf := bytes.NewBufferString("")
	fmt.Println(DecodeAction(buf, newRecvAction(nil)))
}
