package p2p

import (
	"context"
	"errors"
	"glibs/rands"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const helloProbeInterval = 10 * time.Millisecond   // 毫秒
const defaultHolePunchingTimeout = 3 * time.Second // 超时时间
const messageSendRetryNum = 5                      // 最大重试次数

var (
	messageHelloAck = []byte("holePunching:helloAck:1")
	messageSuccess  = []byte("holePunching:success:1")
)

var (
	ErrHolePunchingTimeout = errors.New("hole punching timeout")
)

var logFn = func(format string, v ...interface{}) {}

func SetLog(fn func(format string, v ...interface{})) {
	logFn = fn
}

func wrapError(err error, msg string) error {
	if err == nil {
		return errors.New(msg)
	}
	return errors.New(err.Error() + ", " + msg)
}
func generateDummyPacket() []byte {
	data := rands.GenLetterNumString(rands.GenRangeInt(4, 64))
	return []byte(data)
}

type HolePunchOption func(h *HolePuncher)

func WithProbeInterval(interval time.Duration) HolePunchOption {
	return func(h *HolePuncher) {
		h.proleInterval = interval
	}
}

func WithTimeout(timeout time.Duration) HolePunchOption {
	return func(h *HolePuncher) {
		h.timeout = timeout
	}
}

type HolePuncher struct {
	isStopHello   atomic.Bool
	proleInterval time.Duration
	timeout       time.Duration
}

func NewHolePuncher(opts ...HolePunchOption) *HolePuncher {
	h := &HolePuncher{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (sel *HolePuncher) getTimeout() time.Duration {
	if sel.timeout == 0 {
		return defaultHolePunchingTimeout
	}
	return sel.timeout
}

func (sel *HolePuncher) getProbeInterval() time.Duration {
	if sel.proleInterval == 0 {
		return helloProbeInterval
	}
	return sel.proleInterval
}

func (sel *HolePuncher) HolePunching(localAddr string, remoteAddr string) (*net.UDPConn, error) {
	var wait sync.WaitGroup
	var timeout = sel.getTimeout()

	conn, remoteUdpAddr, err := sel.listenUDP(localAddr, remoteAddr)
	if err != nil {
		return nil, wrapError(err, "listenUDP")
	}

	// 使用 context 统一控制超时和退出
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	wait.Add(1)
	go sel.helloProbeLoop(&wait, ctx, conn, remoteUdpAddr)
	defer wait.Wait()
	defer cancel()

	// 设置读取超时
	if err = conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		_ = conn.Close()
		return nil, wrapError(err, "set read deadline")
	}
	if err = sel.receiveHandle(conn, remoteUdpAddr); err != nil {
		_ = conn.Close()
		return nil, err
	}
	logFn("hole success ", "addr", remoteUdpAddr.String())
	// 取消读取超时
	err = conn.SetReadDeadline(time.Time{})
	return conn, err
}

func (sel *HolePuncher) listenUDP(localAddr string, remoteAddr string) (*net.UDPConn, *net.UDPAddr, error) {
	const protocol = "udp"
	localAddrObj, err := net.ResolveUDPAddr(protocol, localAddr)
	if err != nil {
		return nil, nil, err
	}

	remoteAddrObj, err := net.ResolveUDPAddr(protocol, remoteAddr)
	if err != nil {
		return nil, nil, err
	}

	conn, err := net.ListenUDP(protocol, localAddrObj)
	if err != nil {
		return nil, nil, err
	}
	return conn, remoteAddrObj, nil
}

// helloProbeLoop  sends hello probe to the remote address in loop.
func (sel *HolePuncher) helloProbeLoop(wait *sync.WaitGroup, ctx context.Context, conn *net.UDPConn, remoteAddr *net.UDPAddr) {
	wait.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if sel.isStopHello.Load() {
			return
		}

		msg := generateDummyPacket()
		//logFn("send hello message", "addr", remoteAddr.String(), string(msg))
		if _, err := conn.WriteTo(msg, remoteAddr); err != nil {
			return
		}
		time.Sleep(sel.getProbeInterval())
	}
}

func (sel *HolePuncher) receiveHandle(conn *net.UDPConn, remoteAddr *net.UDPAddr) error {
	for {
		buf := make([]byte, 128)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			// 判断是不是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return ErrHolePunchingTimeout
			}
			return wrapError(err, "read message")
		}
		//logFn("receive message", "addr", addr.String(), "message", string(buf[:n]))
		if remoteAddr.String() != addr.String() {
			continue
		}
		rcvMsg := string(buf[:n])
		switch rcvMsg {
		case string(messageHelloAck):
			if sel.isStopHello.Load() {
				break
			}
			sel.isStopHello.Store(true)
			err = sel.sendPacketWithRetry(conn, remoteAddr, messageSuccess, messageSendRetryNum)
			if err != nil {
				return wrapError(err, "send success message")
			}
		case string(messageSuccess):
			if !sel.isStopHello.Load() {
				sel.isStopHello.Store(true)
				err = sel.sendPacketWithRetry(conn, remoteAddr, messageSuccess, messageSendRetryNum)
				if err != nil {
					return wrapError(err, "send success message")
				}
				return nil
			}
			return nil
		default:
			err = sel.sendPacketWithRetry(conn, remoteAddr, messageHelloAck, messageSendRetryNum)
			if err != nil {
				return wrapError(err, "send helloAck message")
			}
		}
	}

}

func (sel *HolePuncher) sendPacketWithRetry(conn net.PacketConn, remoteAddr *net.UDPAddr, msg []byte, num int) error {
	var err error
	for i := 0; i < num; i++ {
		_, err = conn.WriteTo(msg, remoteAddr)
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 10)
	}
	return nil
}

func HolePunching(localAddr string, remoteAddr string, opts ...HolePunchOption) (*net.UDPConn, error) {
	return NewHolePuncher(opts...).HolePunching(localAddr, remoteAddr)
}
