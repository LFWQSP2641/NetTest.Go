package singdns

import (
	"context"
	"net"
	"time"

	"nettest/pkg/dns/transport"

	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

// dialerAdapter adapts our Dialer/PacketDialer to sing's N.Dialer.
type dialerAdapter struct {
	d  transport.Dialer
	pd transport.PacketDialer
}

func NewDialerAdapter(d transport.Dialer, pd transport.PacketDialer) N.Dialer {
	return &dialerAdapter{d: d, pd: pd}
}

// Expose underlying for internal use by constructors.
func (a *dialerAdapter) OurDialer() transport.Dialer             { return a.d }
func (a *dialerAdapter) OurPacketDialer() transport.PacketDialer { return a.pd }

func (a *dialerAdapter) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if network == N.NetworkUDP || network == "udp" || network == "udp4" || network == "udp6" {
		if a.pd == nil {
			return nil, transport.ErrSocks5UDPUnsupported
		}
		pc, err := a.pd.DialPacket(ctx, "udp", destination.String())
		if err != nil {
			return nil, err
		}
		if c, ok := pc.(net.Conn); ok {
			return c, nil
		}
		return &packetConnAsConn{pc: pc, raddr: destination.UDPAddr()}, nil
	}
	return a.d.DialContext(ctx, network, destination.String())
}

func (a *dialerAdapter) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if a.pd == nil {
		return nil, transport.ErrSocks5UDPUnsupported
	}
	return a.pd.DialPacket(ctx, "udp", destination.String())
}

// packetConnAsConn adapts a PacketConn to a Conn bound to a fixed remote address.
type packetConnAsConn struct {
	pc    net.PacketConn
	raddr net.Addr
}

func (p *packetConnAsConn) Read(b []byte) (int, error) {
	n, _, err := p.pc.ReadFrom(b)
	return n, err
}

func (p *packetConnAsConn) Write(b []byte) (int, error) {
	return p.pc.WriteTo(b, p.raddr)
}

func (p *packetConnAsConn) Close() error                       { return p.pc.Close() }
func (p *packetConnAsConn) LocalAddr() net.Addr                { return p.pc.LocalAddr() }
func (p *packetConnAsConn) RemoteAddr() net.Addr               { return p.raddr }
func (p *packetConnAsConn) SetDeadline(t time.Time) error      { return p.pc.SetDeadline(t) }
func (p *packetConnAsConn) SetReadDeadline(t time.Time) error  { return p.pc.SetReadDeadline(t) }
func (p *packetConnAsConn) SetWriteDeadline(t time.Time) error { return p.pc.SetWriteDeadline(t) }
