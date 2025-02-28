package httpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type VClient struct {
	cltPool *sync.Pool
}

func newVClient() *VClient {
	c := &VClient{
		cltPool: &sync.Pool{
			New: func() interface{} {
				return new(http.Client)
			},
		},
	}
	return c
}

func Do() *VClient {
	if defaultHttpReq == nil {
		defaultHttpReq = newVClient()
	}
	return defaultHttpReq
}

func (vc *VClient) Get(reqUrl string, params map[string]string) (reqBody []byte, err error) {
	if params != nil {
		reqUrl = reqUrl + ParseHttpParamForGet(params)
	}
	client := vc.cltPool.Get().(*http.Client)
	resp, err := client.Get(reqUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	reqBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func (vc *VClient) PostForJson(reqUrl string, params interface{}, head map[string]string) (reqBody []byte, err error) {
	bts, err := jsoniter.Marshal(params)
	if err != nil {
		return
	}
	// TODO 这里可以改为 sync.Pool 对象池
	client := &http.Client{}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(bts))
	req.Close = true
	req.Header.Add("Content-Type", "application/json")

	if head != nil {
		for k, v := range head {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if reqBody, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	return
}

func (vc *VClient) PostXForm(reqUrl string, params, head map[string]string) (reqBody []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(ParseHttpParamForGet(params)))
	req.Close = true
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if head != nil {
		for k, v := range head {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("err = ", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("%d", resp.StatusCode))
	}
	reqBody, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	return
}

func (vc *VClient) PostForJsonWithBaseAuth(reqUrl string, params interface{}, head map[string]string, userName, password string) (reqBody []byte, err error) {
	bts, err := jsoniter.Marshal(params)
	if err != nil {
		return
	}
	// TODO 这里可以改为 sync.Pool 对象池
	client := vc.cltPool.Get().(*http.Client)
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(bts))
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(userName, password)
	if head != nil {
		for k, v := range head {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if reqBody, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	return
}

func (vc *VClient) PostJsonWithTLSCRT(reqUrl string, caFile string, params map[string]interface{}) (reqBody []byte, err error) {

	var buf *bytes.Buffer
	crtPool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(caFile)
	if err != nil {
		return
	}
	crtPool.AppendCertsFromPEM(caCrt)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: crtPool},
	}
	client := vc.cltPool.Get().(*http.Client)
	client.Transport = tr

	if params != nil {
		bts, err := jsoniter.Marshal(params)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(bts)
	}
	req, err := http.NewRequest("POST", reqUrl, buf)
	req.Header.Add("Content-Type", "application/json")
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		return
	}
	defer resp.Body.Close()
	if reqBody, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	return
}

func genCrtHttpClient(crtFile string) (httpC *http.Client, err error) {
	crtPool := x509.NewCertPool()
	crtData, err := ioutil.ReadFile(crtFile)
	if err != nil {
		return
	}
	crtPool.AppendCertsFromPEM(crtData)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: crtPool},
	}
	httpC = &http.Client{
		Transport: tr,
	}
	return
}

func DoRequest(url string, params map[string]interface{}) (resBody []byte, err error) {
	reqBody, err := jsoniter.Marshal(params)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	return doRequest(request)
}

func doRequest(r *http.Request) (resBody []byte, err error) {
	clt, err := genCrtHttpClient("")
	if err != nil {
		return
	}
	resp, err := clt.Do(r)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
