// Definition for the SOAP structure required for UPnP's SOAP usage.

package soap

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	soapEncodingStyle = "http://schemas.xmlsoap.org/soap/encoding/"
	soapPrefix        = xml.Header + `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`
	soapSuffix        = `</s:Body></s:Envelope>`
)

type SOAPClient struct {
	controlURL url.URL
	httpClient http.Client
}

func NewSOAPClient(controlURL url.URL) *SOAPClient {
	return &SOAPClient{
		controlURL: controlURL,
	}
}

// DoCallCtx makes a SOAP request, with the given action.
// inAction and outAction must both be pointers to structs with string fields only.
func (client *SOAPClient) DoCallCtx(ctx context.Context, actionNamespace, actionName string, inAction interface{}, outAction interface{}) error {
	reqBuf := bytes.NewBuffer(nil)
	action := newSendAction(actionNamespace, inAction, actionName)
	if err := EncodeAction(reqBuf, action); err != nil {
		return err
	}

	//fmt.Println("Namespace：", action.Namespace())
	//fmt.Println("ActionName：", action.ActionName())
	fmt.Println("获取请求值：", reqBuf.String())

	req := &http.Request{
		Method:        "POST",
		URL:           &client.controlURL,
		Body:          io.NopCloser(reqBuf),
		ContentLength: int64(reqBuf.Len()),
		Header: http.Header{
			"SOAPAction":   []string{`"` + action.Namespace() + "#" + action.ActionName() + `"`},
			"Content-Type": []string{"text/xml; charset=\"utf-8\""},
		},
	}

	response, err := client.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("http do request fault: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.ContentLength == 0 {
		return fmt.Errorf("soap http request fault: %s", response.Status)
	}

	responseBody, _ := io.ReadAll(response.Body)
	fmt.Println("获取响应值：", string(responseBody))

	if err = DecodeAction(bytes.NewBuffer(responseBody), newRecvAction(outAction)); err != nil {
		return fmt.Errorf("decode response fault: %v", err)
	}

	return nil
}

// DoCall is the legacy version of DoCallCtx, which uses context.Background.
func (client *SOAPClient) DoCall(actionNamespace, actionName string, inAction interface{}, outAction interface{}) error {
	return client.DoCallCtx(context.Background(), actionNamespace, actionName, inAction, outAction)
}
