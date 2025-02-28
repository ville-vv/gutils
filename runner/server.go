package runner

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server interface {
	Start()
	Stop()
}

type ServerGroup struct {
	servers  []Server
	onceStop sync.Once
}

func NewServerGroup() *ServerGroup {
	sg := &ServerGroup{
		servers:  make([]Server, 0),
		onceStop: sync.Once{},
	}
	return sg
}

func (sel *ServerGroup) Start() {
	// 创建一个信号通道
	signalChan := make(chan os.Signal, 1)
	// 监控系统信号，如中断（Ctrl+C）和终止信号
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-signalChan
	sel.Stop()
	time.Sleep(time.Millisecond * 100)
}

func (sel *ServerGroup) start() {
	var wg sync.WaitGroup
	for i := range sel.servers {
		svr := sel.servers[i]
		wg.Add(1)
		go func(inWg *sync.WaitGroup) {
			inWg.Done()
			svr.Start()
		}(&wg)
	}
	wg.Wait()
}

func (sel *ServerGroup) Stop() {
	sel.onceStop.Do(func() {
		for _, svr := range sel.servers {
			svr.Stop()
		}
	})
}
