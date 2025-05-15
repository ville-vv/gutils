package nat

import (
	"context"
	"i4remoter/pkg/nat/ssdp"
	upnp2 "i4remoter/pkg/nat/upnp"
	"i4remoter/utils/zlog"
	"net/url"
	"sync"
	"time"
)

var _discover = &Discover{}

type Discover struct {
}

func DiscoverNats(ctx context.Context) ([]NAT, error) {
	return _discover.DiscoverNats(ctx)
}

func (sel *Discover) discoverAsync(ctx context.Context, wait *sync.WaitGroup, natsCha chan []NAT, discoverFun func(ctx context.Context) ([]NAT, error)) {
	wait.Add(1)
	go func() {
		defer wait.Done()
		defer func() {
			if err := recover(); err != nil {
				_log.Errorf("async discover recover %v", err)
			}
		}()
		nats, err := discoverFun(ctx)
		if err != nil {
			_log.Errorf("async discover error %s", err.Error())
			return
		}
		natsCha <- nats
	}()
}

func (sel *Discover) DiscoverNats(ctx context.Context) ([]NAT, error) {
	var wait sync.WaitGroup
	var natsCh = make(chan []NAT, 2)
	var backNatsCh = make(chan []NAT, 1)
	sel.discoverAsync(ctx, &wait, natsCh, sel.discoverUPNPIG1)
	sel.discoverAsync(ctx, &wait, natsCh, sel.discoverUPNPIG2)
	sel.discoverAsync(ctx, &wait, backNatsCh, sel.discoverOtherIG)
	go func() {
		wait.Wait()
		close(natsCh)
	}()
	outNats := make([]NAT, 0)
	for nats := range natsCh {
		outNats = append(outNats, nats...)
	}
	defer close(backNatsCh)
	if len(outNats) == 0 {
		select {
		case <-ctx.Done():
		case nats, ok := <-backNatsCh:
			if !ok {
				break
			}
			outNats = append(outNats, nats...)
		case <-time.After(time.Millisecond * 300):
			break
		}
	}
	return outNats, nil
}

func (sel *Discover) discoverUPNPIG1(ctx context.Context) ([]NAT, error) {
	return sel.searchWithType(ctx, upnp2.URN_DEVICE_WANConnectionDevice_1)
}

func (sel *Discover) discoverUPNPIG2(ctx context.Context) ([]NAT, error) {
	return sel.searchWithType(ctx, upnp2.URN_DEVICE_WANConnectionDevice_2)
}

func (sel *Discover) searchWithType(ctx context.Context, searchType string) ([]NAT, error) {
	deviceList, err := ssdp.Search(ctx, searchType)
	if err != nil {
		return nil, err
	}
	nats := make([]NAT, 0, len(deviceList))
	for _, dev := range deviceList {
		sel.visitServices(ctx, dev.LOCATION, &nats)
	}
	return nats, nil
}

func (sel *Discover) visitServices(ctx context.Context, loc *url.URL, nats *[]NAT) error {
	rootDevice, err := upnp2.DeviceByURLCtx(ctx, loc)
	if err != nil {
		return err
	}
	rootDevice.VisitServices(collectNATServices(ctx, rootDevice, nats))
	return nil
}

// discoverOtherIG is a fallback for routers that fail to respond to our
// SSDP queries. It will query all devices and try to find any InternetGatewayDevice.
func (sel *Discover) discoverOtherIG(ctx context.Context) ([]NAT, error) {
	serviceList, err := ssdp.Search(ctx, ssdp.SSDPAll)
	if err != nil {
		return nil, err
	}

	locationMap := make(map[string]struct{})

	nats := make([]NAT, 0, len(serviceList))
	for _, svc := range serviceList {
		zlog.Infow("搜索到设备", "location", svc.LOCATION.String(), "st", svc.ST)
		if upnp2.IsInternalGatewayDevice(svc.ST) {
			continue
		}
		// 同样的 location 只处理一次
		locStr := svc.LOCATION.String()
		if _, ok := locationMap[locStr]; ok {
			continue
		}
		locationMap[locStr] = struct{}{}
		sel.visitServices(ctx, svc.LOCATION, &nats)
	}
	return nats, nil
}

// collectNATServices is a vistor function that visits all services of a root
func collectNATServices(ctx context.Context, rootDevice *upnp2.RootDevice, outNats *[]NAT) func(srv *upnp2.Service) {
	return func(srv *upnp2.Service) {
		if ctx.Err() != nil {
			return
		}
		switch srv.ServiceType {
		case upnp2.URN_SERVICE_WANIPConnection_1:
			client := upnp2.NewWANIPConnection1(srv.NewSOAPClient())
			*outNats = append(*outNats, &upnpNat{
				upnpCli:    client,
				natType:    "UPNP(IP1)",
				location:   rootDevice.URLBase,
				deviceName: rootDevice.Device.FriendlyName,
			})
		case upnp2.URN_SERVICE_WANIPConnection_2:
			if rootDevice.Device.DeviceType == upnp2.URN_DEVICE_WANConnectionDevice_1 {
				// found V2 service on V1 device
				return
			}
			client := upnp2.NewWANPPPConnection1(srv.NewSOAPClient())
			*outNats = append(*outNats, &upnpNat{
				upnpCli:    client,
				natType:    "UPNP(IP2)",
				location:   rootDevice.URLBase,
				deviceName: rootDevice.Device.FriendlyName,
			})
		case upnp2.URN_SERVICE_WANPPPConnection_1:
			client := upnp2.NewWANPPPConnection1(srv.NewSOAPClient())
			*outNats = append(*outNats, &upnpNat{
				upnpCli:    client,
				natType:    "UPNP(PPP1)",
				location:   rootDevice.URLBase,
				deviceName: rootDevice.Device.FriendlyName,
			})
		}
	}
}
