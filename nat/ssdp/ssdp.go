package ssdp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ssdpDiscover   = `"ssdp:discover"`
	ntsAlive       = `ssdp:alive`
	ntsByeBye      = `ssdp:byebye`
	ntsUpdate      = `ssdp:update`
	ssdpUDP4Addr   = "239.255.255.250:1900"
	ssdpSearchPort = 1900
	methodSearch   = "M-SEARCH"
	methodNotify   = "NOTIFY"
	SSDPAll        = "ssdp:all"        // SSDPAll is a value for searchTarget that searches for all devices and services.
	UPNPRootDevice = "upnp:rootdevice" // UPNPRootDevice is a value for searchTarget that searches for all root devices.
)

func ssdpIp4NetAddr() net.Addr {
	addr, _ := net.ResolveUDPAddr("udp4", ssdpUDP4Addr)
	return addr
}

func buildMessage(req *http.Request) ([]byte, error) {
	var requestBuf bytes.Buffer
	method := req.Method
	if method == "" {
		method = "GET"
	}
	if _, err := fmt.Fprintf(&requestBuf, "%s %s HTTP/1.1\r\n", method, req.URL.RequestURI()); err != nil {
		return nil, err
	}
	if err := req.Header.Write(&requestBuf); err != nil {
		return nil, err
	}
	if _, err := requestBuf.Write([]byte{'\r', '\n'}); err != nil {
		return nil, err
	}
	return requestBuf.Bytes(), nil
}

type Service struct {
	ST       string
	USN      string
	SERVER   string
	LOCATION *url.URL
}

func (sel *Service) String() string {
	return fmt.Sprintf("ST: %s, USN: %s, SERVER: %s, LOCATION: %s", sel.ST, sel.USN, sel.SERVER, sel.LOCATION.String())
}

type IService interface {
	Location() (*url.URL, error)
	Server() string
	USN() string
	ST() string
}

type Response struct {
	*http.Response
}

func (sel *Response) ST() string {
	return sel.Header.Get("ST")
}

func (sel *Response) USN() string {
	return sel.Header.Get("USN")
}

func (sel *Response) Server() string {
	return sel.Header.Get("SERVER")
}

func SearchWithTimeout(ctx context.Context, searchType string, maxWaitSeconds int) ([]Service, error) {
	destAddr := ssdpIp4NetAddr()
	conn, err := NewConn(destAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req, err := newRequest(ctx, searchType, maxWaitSeconds)
	if err != nil {
		return nil, err
	}
	response, err := conn.DoRequest(req, time.Millisecond*time.Duration(maxWaitSeconds*100))
	if err != nil {
		return nil, err
	}
	return processSSDPResponses(searchType, response)
}

func Search(ctx context.Context, searchType string) ([]Service, error) {
	return SearchWithTimeout(ctx, searchType, 1)
}

func newRequest(ctx context.Context, serviceType string, maxWaitSeconds int) (*http.Request, error) {
	if maxWaitSeconds < 1 {
		return nil, errors.New("ssdp: request timeout must be at least 1s")
	}
	req := (&http.Request{
		Method: methodSearch,
		Host:   ssdpUDP4Addr,
		URL:    &url.URL{Opaque: "*"},
		Header: http.Header{
			"HOST": []string{ssdpUDP4Addr},
			"MX":   []string{strconv.Itoa(maxWaitSeconds)},
			"MAN":  []string{ssdpDiscover},
			"ST":   []string{serviceType},
		},
	}).WithContext(ctx)

	return req, nil
}

func processSSDPResponses(serviceType string, allResponses []*http.Response) ([]Service, error) {
	isExactSearch := serviceType != SSDPAll && serviceType != UPNPRootDevice
	seenIDs := make(map[string]bool)
	var responses []Service
	for _, val := range allResponses {
		response := &Response{val}
		if response.StatusCode != 200 {
			continue
		}
		if st := response.ST(); isExactSearch && st != serviceType {
			continue
		}

		loc, err := response.Location()
		if err != nil {
			continue
		}

		id := loc.String() + "\x00" + response.USN()
		if _, ok := seenIDs[id]; !ok {
			seenIDs[id] = true
			responses = append(responses, Service{
				ST:       response.ST(),
				USN:      response.USN(),
				SERVER:   response.Server(),
				LOCATION: loc,
			})
		}
	}
	return responses, nil
}
