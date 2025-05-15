package main

import (
	"context"
	"flag"
	"fmt"
	"i4remoter/pkg/nat"
)

func main() {
	var internalPort int
	var leaseDuration uint
	var externalPort int
	var help bool
	flag.IntVar(&internalPort, "inPort", 17269, "internalPort")
	flag.IntVar(&externalPort, "exPort", 17269, "externalPort")
	flag.UintVar(&leaseDuration, "d", 3600, "leaseDuration")
	flag.BoolVar(&help, "h", false, "help")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	natList, err := nat.DiscoverNats(context.Background())
	if err != nil {
		fmt.Println("错误：", err)
		return
	}

	for i, dev := range natList {
		fmt.Println("请求Desc地址：", i, dev.Type(), dev.DeviceName(), dev.Location().String())
		fmt.Println(dev.GetExternalAddress())
		fmt.Println(dev.GetInternalAddress())
		fmt.Println(dev.GetDeviceAddress())
		fmt.Println(dev.GetDeviceStatus())
		err = dev.AddPortMapping("udp", internalPort, externalPort, leaseDuration, "remoteP2P")
		if err != nil {
			fmt.Println("错误：", err)
			return
		}
		fmt.Println(dev.GetPortMapping("udp", externalPort))
	}
}
