package ssdp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/ipv4"
)

type connConfig struct {
	ttl   int
	sysIf bool
}

// NewConn starts to receiving multicast messages.
func NewConn(destAddr net.Addr, opts ...ConnOption) (*Conn, error) {
	var cfg connConfig
	for _, o := range opts {
		o.apply(&cfg)
	}

	pconn, ifplist, err := newIPv4MulticastConn(cfg.ttl, destAddr)
	if err != nil {
		return nil, err
	}

	return &Conn{
		raddr: destAddr,
		laddr: pconn.LocalAddr(),
		pconn: pconn,
		ifps:  ifplist,
	}, nil
}

// newIPv4MulticastConn create a new multicast connection.
// 2nd return parameter will be nil when sysIf is true.
func newIPv4MulticastConn(ttl int, destAddr net.Addr) (*ipv4.PacketConn, []*net.Interface, error) {
	laddr, err := net.ResolveUDPAddr("udp4", "")
	if err != nil {
		return nil, nil, err
	}

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return nil, nil, err
	}

	list, _ := GetInterfacesIPv4()
	pConn, err := joinGroupIPv4(conn, list, destAddr)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if ttl > 0 {
		if err = pConn.SetTTL(ttl); err != nil {
			pConn.Close()
			return nil, nil, err
		}
	}
	return pConn, list, nil
}

// joinGroupIPv4 makes the connection join to a group on interfaces.
// This trys to use system assigned when iflist is nil or empty.
func joinGroupIPv4(conn *net.UDPConn, list []*net.Interface, gaddr net.Addr) (*ipv4.PacketConn, error) {
	wrap := ipv4.NewPacketConn(conn)
	wrap.SetMulticastLoopback(true)
	if len(list) == 0 {
		if err := wrap.JoinGroup(nil, gaddr); err != nil {
			return nil, errors.New("no system assigned multicast interfaces had joined to group")
		}
		return wrap, nil
	}
	// add interfaces to multicast group.
	joined := 0
	for _, ifi := range list {
		if err := wrap.JoinGroup(ifi, gaddr); err != nil {
			continue
		}
		joined++
	}
	if joined == 0 {
		return nil, errors.New("no interfaces had joined to group")
	}
	return wrap, nil
}

type Conn struct {
	raddr net.Addr
	laddr net.Addr
	pconn *ipv4.PacketConn
	ifps  []*net.Interface // ifps stores pointers of multicast interface.
}

// Close closes a multicast connection.
func (sel *Conn) Close() error {
	if err := sel.pconn.Close(); err != nil {
		return err
	}
	// based net.UDPConn will be closed by sel.pconn.Close()
	return nil
}

// WriteTo sends a multicast message to interfaces.
func (sel *Conn) WriteTo(data []byte, to net.Addr) (int, error) {
	// Send a multicast message directory when recipient "to" address is not multicast.
	if uaddr, ok := to.(*net.UDPAddr); ok && !uaddr.IP.IsMulticast() {
		return sel.writeToIfi(data, to, nil)
	}
	// Send a multicast message to all interfaces (iflist).
	sum := 0
	for _, ifi := range sel.ifps {
		n, err := sel.writeToIfi(data, to, ifi)
		if err != nil {
			return 0, err
		}
		sum += n
	}
	return sum, nil
}

func (sel *Conn) DoRequest(req *http.Request, timeout time.Duration) ([]*http.Response, error) {
	msg, err := buildMessage(req)
	if err != nil {
		return nil, err
	}

	if _, err = sel.WriteTo(msg, sel.raddr); err != nil {
		return nil, err
	}
	var responses []*http.Response
	err = sel.ReadPackets(timeout, func(addr net.Addr, data []byte) error {
		resp, inErr := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
		if inErr != nil {
			return err
		}
		responses = append(responses, resp)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return responses, nil
}

func (sel *Conn) writeToIfi(data []byte, to net.Addr, ifi *net.Interface) (int, error) {
	if err := sel.pconn.SetMulticastInterface(ifi); err != nil {
		return 0, err
	}
	return sel.pconn.WriteTo(data, nil, to)
}

// LocalAddr returns local address to listen multicast packets.
func (sel *Conn) LocalAddr() net.Addr {
	return sel.laddr
}

type PacketHandler func(addr net.Addr, data []byte) error

// ReadPackets reads multicast packets.
func (sel *Conn) ReadPackets(timeout time.Duration, h PacketHandler) error {
	buf := make([]byte, 65535)
	if timeout > 0 {
		sel.pconn.SetReadDeadline(time.Now().Add(timeout))
	}
	for {
		n, _, addr, err := sel.pconn.ReadFrom(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				return nil
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return io.EOF
			}
			return err
		}
		if err = h(addr, buf[:n]); err != nil {
			return err
		}
	}
}

func (sel *Conn) parseData(data []byte) (*http.Response, error) {
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ConnOption is option for Listen()
type ConnOption interface {
	apply(cfg *connConfig)
}

type connOptFunc func(*connConfig)

func (f connOptFunc) apply(cfg *connConfig) {
	f(cfg)
}

func ConnTTL(ttl int) ConnOption {
	return connOptFunc(func(cfg *connConfig) {
		cfg.ttl = ttl
	})
}

func ConnSystemAssginedInterface() ConnOption {
	return connOptFunc(func(cfg *connConfig) {
		cfg.sysIf = true
	})
}
