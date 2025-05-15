package upnp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"i4remoter/pkg/nat/scpd"
	"i4remoter/pkg/nat/soap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	deviceXMLNamespace = "urn:schemas-upnp-org:device-1-0"
)

// RootDevice is the device description as described by section 2.3 "Device
// description" in
// http://upnp.org/specs/arch/UPnP-arch-DeviceArchitecture-v1.1.pdf
type RootDevice struct {
	XMLName     xml.Name    `xml:"root"`
	SpecVersion SpecVersion `xml:"specVersion"`
	URLBase     url.URL     `xml:"-"`
	URLBaseStr  string      `xml:"URLBase"`
	Device      Device      `xml:"device"`
}

// SetURLBase sets the URLBase for the RootDevice and its underlying components.
func (root *RootDevice) SetURLBase(urlBase *url.URL) {
	root.URLBase = *urlBase
	root.URLBaseStr = urlBase.String()
	root.Device.SetURLBase(urlBase)
}

func (root *RootDevice) VisitServices(visitor func(*Service)) {
	root.Device.VisitServices(visitor)
}

func (root *RootDevice) VisitDevices(visitor func(*Device)) {
	root.Device.VisitDevices(visitor)
}

// SpecVersion is part of a RootDevice, describes the version of the
// specification that the data adheres to.
type SpecVersion struct {
	Major int32 `xml:"major"`
	Minor int32 `xml:"minor"`
}

// Device is a UPnP device. It can have child devices.
type Device struct {
	DeviceType       string    `xml:"deviceType"`
	FriendlyName     string    `xml:"friendlyName"`
	Manufacturer     string    `xml:"manufacturer"`
	ManufacturerURL  URLField  `xml:"manufacturerURL"`
	ModelDescription string    `xml:"modelDescription"`
	ModelName        string    `xml:"modelName"`
	ModelNumber      string    `xml:"modelNumber"`
	ModelType        string    `xml:"modelType"`
	ModelURL         URLField  `xml:"modelURL"`
	SerialNumber     string    `xml:"serialNumber"`
	UDN              string    `xml:"UDN"`
	UPC              string    `xml:"UPC,omitempty"`
	Icons            []Icon    `xml:"iconList>icon,omitempty"`
	Services         []Service `xml:"serviceList>service,omitempty"`
	Devices          []Device  `xml:"deviceList>device,omitempty"`
	PresentationURL  URLField  `xml:"presentationURL"` // Extra observed elements:
}

// VisitDevices calls visitor for the device, and all its descendent devices.
func (sel *Device) VisitDevices(visitor func(*Device)) {
	visitor(sel)
	for i := range sel.Devices {
		sel.Devices[i].VisitDevices(visitor)
	}
}

// VisitServices calls visitor for all Services under the device and all its
// descendent devices.
func (sel *Device) VisitServices(visitor func(*Service)) {
	sel.VisitDevices(func(d *Device) {
		if d == nil {
			return
		}
		for i := range d.Services {
			visitor(&d.Services[i])
		}
	})
}

// FindService finds all (if any) Services under the device and its descendents
// that have the given ServiceType.
func (sel *Device) FindService(serviceType string) []*Service {
	var services []*Service
	sel.VisitServices(func(s *Service) {
		if s.ServiceType == serviceType {
			services = append(services, s)
		}
	})
	return services
}

// SetURLBase sets the URLBase for the Device and its underlying components.
func (sel *Device) SetURLBase(urlBase *url.URL) {
	sel.ManufacturerURL.SetURLBase(urlBase)
	sel.ModelURL.SetURLBase(urlBase)
	sel.PresentationURL.SetURLBase(urlBase)
	for i := range sel.Icons {
		sel.Icons[i].SetURLBase(urlBase)
	}
	for i := range sel.Services {
		sel.Services[i].SetURLBase(urlBase)
	}
	for i := range sel.Devices {
		sel.Devices[i].SetURLBase(urlBase)
	}
}

func (sel *Device) String() string {
	return fmt.Sprintf("Device ID %s : %s (%s)", sel.UDN, sel.DeviceType, sel.FriendlyName)
}

// Icon is a representative image that a device might include in its
// description.
type Icon struct {
	Mimetype string   `xml:"mimetype"`
	Width    int32    `xml:"width"`
	Height   int32    `xml:"height"`
	Depth    int32    `xml:"depth"`
	URL      URLField `xml:"url"`
}

// SetURLBase sets the URLBase for the Icon.
func (icon *Icon) SetURLBase(url *url.URL) {
	icon.URL.SetURLBase(url)
}

// Service is a service provided by a UPnP Device.
type Service struct {
	ServiceType string   `xml:"serviceType"`
	ServiceId   string   `xml:"serviceId"`
	SCPDURL     URLField `xml:"SCPDURL"`
	ControlURL  URLField `xml:"controlURL"`
	EventSubURL URLField `xml:"eventSubURL"`
}

// SetURLBase sets the URLBase for the Service.
func (srv *Service) SetURLBase(urlBase *url.URL) {
	srv.SCPDURL.SetURLBase(urlBase)
	srv.ControlURL.SetURLBase(urlBase)
	srv.EventSubURL.SetURLBase(urlBase)
}

func (srv *Service) String() string {
	return fmt.Sprintf("Service ID %s : %s", srv.ServiceId, srv.ServiceType)
}

// RequestSCPDCtx requests the SCPD (soap actions and state variables description)
// for the service.
func (srv *Service) RequestSCPDCtx(ctx context.Context) (*scpd.SCPD, error) {
	if !srv.SCPDURL.Ok {
		return nil, errors.New("bad/missing SCPD URL, or no URLBase has been set")
	}
	s := new(scpd.SCPD)
	if err := requestXml(ctx, srv.SCPDURL.URL.String(), scpd.SCPDXMLNamespace, s); err != nil {
		return nil, err
	}
	return s, nil
}

// RequestSCPD is the legacy version of RequestSCPDCtx, but uses
// context.Background() as the context.
func (srv *Service) RequestSCPD() (*scpd.SCPD, error) {
	return srv.RequestSCPDCtx(context.Background())
}

// RequestSCDP is for compatibility only, prefer RequestSCPD. This was a
// misspelling of RequestSCDP.
func (srv *Service) RequestSCDP() (*scpd.SCPD, error) {
	return srv.RequestSCPD()
}

func (srv *Service) NewSOAPClient() *soap.SOAPClient {
	return soap.NewSOAPClient(srv.ControlURL.URL)
}

// URLField is a URL that is part of a device description.
type URLField struct {
	URL url.URL `xml:"-"`
	Ok  bool    `xml:"-"`
	Str string  `xml:",chardata"`
}

func (uf *URLField) SetURLBase(urlBase *url.URL) {
	str := uf.Str
	if !strings.Contains(str, "://") && !strings.HasPrefix(str, "/") {
		str = "/" + str
	}

	refUrl, err := url.Parse(str)
	if err != nil {
		uf.URL = url.URL{}
		uf.Ok = false
		return
	}

	uf.URL = *urlBase.ResolveReference(refUrl)
	uf.Ok = true
}

var HTTPClientDefault = http.DefaultClient
var CharsetReaderDefault func(charset string, input io.Reader) (io.Reader, error)

func requestXml(ctx context.Context, url string, defaultSpace string, doc interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := HTTPClientDefault.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("got response status %s from %q", resp.Status, url)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.DefaultSpace = defaultSpace
	decoder.CharsetReader = CharsetReaderDefault

	return decoder.Decode(doc)
}

func DeviceByURLCtx(ctx context.Context, loc *url.URL) (*RootDevice, error) {
	locStr := loc.String()
	root := new(RootDevice)
	if err := requestXml(ctx, locStr, deviceXMLNamespace, root); err != nil {
		return nil, err
	}
	var urlBaseStr string
	if root.URLBaseStr != "" {
		urlBaseStr = root.URLBaseStr
	} else {
		urlBaseStr = locStr
	}
	urlBase, err := url.Parse(urlBaseStr)
	if err != nil {
		return nil, err
	}
	root.SetURLBase(urlBase)
	return root, nil
}
