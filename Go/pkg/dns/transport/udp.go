package transport

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/miekg/dns"
)

const defaultUDPTimeout = 5 * time.Second

// ExchangeUDPWithDialer sends a DNS message over UDP using the provided PacketDialer.
// - server: "host:port" (e.g., "8.8.8.8:53")
// - d: a PacketDialer, e.g., DirectDialer or Socks5Dialer
func ExchangeUDPWithDialer(ctx context.Context, msg *dns.Msg, server string, d PacketDialer) (*dns.Msg, time.Duration, error) {
	if msg == nil {
		return nil, 0, errors.New("nil dns msg")
	}
	if server == "" {
		return nil, 0, errors.New("empty server")
	}

	// Pack query
	payload, err := msg.Pack()
	if err != nil {
		return nil, 0, err
	}

	// Dial UDP PacketConn via pluggable dialer
	pc, err := d.DialPacket(ctx, "udp", server)
	if err != nil {
		return nil, 0, err
	}
	defer pc.Close()

	// Apply deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		_ = pc.SetDeadline(deadline)
	}

	start := time.Now()

	// Write
	if uc, ok := pc.(*net.UDPConn); ok {
		// Connected UDP: use Write
		if _, err = uc.Write(payload); err != nil {
			return nil, 0, err
		}
	} else {
		// Generic PacketConn: WriteTo with fixed remote
		if _, err = pc.WriteTo(payload, nil); err != nil {
			return nil, 0, err
		}
	}

	// Read response
	// Typical DNS over UDP size; EDNS may return larger but 4096 is common and safe
	buf := make([]byte, 4096)
	var n int
	if uc, ok := pc.(*net.UDPConn); ok {
		n, err = uc.Read(buf)
	} else {
		n, _, err = pc.ReadFrom(buf)
	}
	if err != nil {
		return nil, 0, err
	}

	rtt := time.Since(start)

	// Unpack response
	resp := new(dns.Msg)
	if err := resp.Unpack(buf[:n]); err != nil {
		return nil, 0, err
	}
	return resp, rtt, nil
}

// ExchangeUDP is a convenience wrapper for direct UDP without a proxy.
// It uses a DirectDialer with a default timeout and a context with the same timeout.
func ExchangeUDP(msg *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultUDPTimeout)
	defer cancel()
	d := NewDirectDialer(DialOptions{Timeout: defaultUDPTimeout})
	return ExchangeUDPWithDialer(ctx, msg, server, d)
}
