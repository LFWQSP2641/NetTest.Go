package transport

import (
	"context"
	"errors"
	"net"
	"time"

	s5 "github.com/txthinking/socks5"
)

var (
	ErrSocks5UDPUnsupported = errors.New("socks5 UDP associate unsupported")
	ErrPacketConnAddrLocked = errors.New("packet conn is locked to a fixed remote address")
)

type Socks5Dialer struct {
	ProxyAddr string
	Username  string
	Password  string
	Timeout   time.Duration
	KeepAlive time.Duration
}

func NewSocks5Dialer(addr, user, pass string, opts DialOptions) *Socks5Dialer {
	return &Socks5Dialer{
		ProxyAddr: addr,
		Username:  user,
		Password:  pass,
		Timeout:   opts.Timeout,
		KeepAlive: opts.KeepAlive,
	}
}

// DialContext: 通过 SOCKS5 CONNECT 拨 TCP（用于 TCP/TLS/HTTP）
func (s *Socks5Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, errors.New("socks5 dial supports only tcp/tcp4/tcp6")
	}
	tcpTO := int(s.Timeout / time.Second)
	udpTO := tcpTO
	c, err := s5.NewClient(s.ProxyAddr, s.Username, s.Password, tcpTO, udpTO)
	if err != nil {
		return nil, err
	}
	// 通过库的 Dial 建立到目标的流式连接
	conn, err := c.Dial(network, address)
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	// 应用 ctx 截止时间
	if deadline, ok := ctx.Deadline(); ok {
		// Set deadline on the established connection; avoid calling Client.SetDeadline (may not be safe before negotiate)
		_ = conn.SetDeadline(deadline)
	}
	// 注意：c 的生命周期与 conn 绑定，conn 关闭时再关闭 c
	return &socks5Conn{Conn: conn, client: c}, nil
}

// DialPacket: 通过 SOCKS5 UDP（单目标地址包装成 net.PacketConn）
func (s *Socks5Dialer) DialPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, errors.New("socks5 UDP only supports udp/udp4/udp6")
	}
	tcpTO := int(s.Timeout / time.Second)
	udpTO := tcpTO
	c, err := s5.NewClient(s.ProxyAddr, s.Username, s.Password, tcpTO, udpTO)
	if err != nil {
		return nil, err
	}
	// 通过库的 UDP 拨号建立“固定目标”的 UDP 交换连接
	uc, err := c.Dial("udp", address)
	if err != nil {
		_ = c.Close()
		return nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = uc.SetDeadline(deadline)
	}
	// 将目标地址解析为 net.Addr，封装成 net.PacketConn（单目标）
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		_ = uc.Close()
		_ = c.Close()
		return nil, err
	}
	return &s5PacketConn{
		client:     c,
		conn:       uc,    // 库返回的 UDP“连接”，内部已处理 SOCKS5 UDP 封装
		remoteAddr: raddr, // 固定远端
	}, nil
}

// socks5Conn 让 client 生命周期与 conn 绑定，Close 时一并关闭
type socks5Conn struct {
	// keep net.Conn embedded to satisfy net.Conn; hold Client as a named field to avoid method conflicts
	net.Conn
	client *s5.Client
}

func (c *socks5Conn) Close() error {
	_ = c.Conn.Close()
	return c.client.Close()
}

// s5PacketConn: 单目标 UDP 的 net.PacketConn 适配器
type s5PacketConn struct {
	// do not embed Client/Conn anonymously to avoid method name collisions
	client     *s5.Client
	conn       net.Conn // 库侧 UDP“连接”，Read/Write 为纯负载
	remoteAddr net.Addr
}

func (pc *s5PacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := pc.conn.Read(b)
	return n, pc.remoteAddr, err
}

func (pc *s5PacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	// 仅允许写到固定远端
	if addr == nil {
		addr = pc.remoteAddr
	}
	if addr.String() != pc.remoteAddr.String() {
		return 0, ErrPacketConnAddrLocked
	}
	return pc.conn.Write(b)
}

func (pc *s5PacketConn) Close() error {
	_ = pc.conn.Close()
	return pc.client.Close()
}

func (pc *s5PacketConn) LocalAddr() net.Addr { return pc.conn.LocalAddr() }

func (pc *s5PacketConn) SetDeadline(t time.Time) error { return pc.conn.SetDeadline(t) }

func (pc *s5PacketConn) SetReadDeadline(t time.Time) error { return pc.conn.SetReadDeadline(t) }

func (pc *s5PacketConn) SetWriteDeadline(t time.Time) error { return pc.conn.SetWriteDeadline(t) }
