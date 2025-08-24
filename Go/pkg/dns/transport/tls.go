package transport

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/miekg/dns"
)

const defaultTLSTimeout = 7 * time.Second

// ExchangeTLSWithDialer sends a DNS message over TLS (DoT, RFC 7858) using the provided Dialer.
// It uses TCP length-prefixed framing over a TLS stream.
// - server: "host:port" (e.g., "1.1.1.1:853" or "dns.google:853")
// - d: a Dialer, e.g., DirectDialer or Socks5Dialer
func ExchangeTLSWithDialer(ctx context.Context, msg *dns.Msg, server string, d Dialer, to TLSOptions) (*dns.Msg, time.Duration, error) {
	if msg == nil {
		return nil, 0, errors.New("nil dns msg")
	}
	if server == "" {
		return nil, 0, errors.New("empty server")
	}

	payload, err := msg.Pack()
	if err != nil {
		return nil, 0, err
	}
	if len(payload) > 0xFFFF {
		return nil, 0, errors.New("dns tcp payload too large")
	}

	// 1) TCP via pluggable dialer
	baseConn, err := d.DialContext(ctx, "tcp", server)
	if err != nil {
		return nil, 0, err
	}
	// Ensure base conn is closed on all paths
	// tlsConn.Close() will also close underlying conn, but keep a defer just in case of handshake failure before wrapping
	defer baseConn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = baseConn.SetDeadline(deadline)
	}

	// 2) TLS client over the base connection
	cfg := &tls.Config{
		ServerName:         to.ServerName,
		InsecureSkipVerify: to.InsecureSkipVerify,
		RootCAs:            to.RootCAs,
	}
	if len(to.NextProtos) > 0 {
		cfg.NextProtos = to.NextProtos
	}
	tlsConn := tls.Client(baseConn, cfg)
	// Apply deadline to TLS layer as well
	if deadline, ok := ctx.Deadline(); ok {
		_ = tlsConn.SetDeadline(deadline)
	}
	if err := tlsConn.Handshake(); err != nil {
		_ = tlsConn.Close()
		return nil, 0, err
	}
	// After handshake, deadlines still apply to the underlying conn

	start := time.Now()

	// 3) Write length-prefixed query
	var lp [2]byte
	binary.BigEndian.PutUint16(lp[:], uint16(len(payload)))
	if err := writeFull(tlsConn, lp[:]); err != nil {
		_ = tlsConn.Close()
		return nil, 0, err
	}
	if err := writeFull(tlsConn, payload); err != nil {
		_ = tlsConn.Close()
		return nil, 0, err
	}

	// 4) Read length-prefixed response
	var hp [2]byte
	if _, err := io.ReadFull(tlsConn, hp[:]); err != nil {
		_ = tlsConn.Close()
		return nil, 0, err
	}
	n := int(binary.BigEndian.Uint16(hp[:]))
	if n <= 0 {
		_ = tlsConn.Close()
		return nil, 0, errors.New("empty dns tls response")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(tlsConn, buf); err != nil {
		_ = tlsConn.Close()
		return nil, 0, err
	}

	rtt := time.Since(start)
	_ = tlsConn.Close()

	resp := new(dns.Msg)
	if err := resp.Unpack(buf); err != nil {
		return nil, 0, err
	}
	return resp, rtt, nil
}

// ExchangeTLS is a convenience wrapper for direct DoT without a proxy.
// Note: if server is an IP (e.g., 1.1.1.1:853), certificate validation will typically fail
// unless you set TLSOptions.ServerName to the certificate name (e.g., cloudflare-dns.com) or
// InsecureSkipVerify to true. Here we default to InsecureSkipVerify=true to keep the helper usable.
func ExchangeTLS(msg *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTLSTimeout)
	defer cancel()
	d := NewDirectDialer(DialOptions{Timeout: defaultTLSTimeout})
	// Default to insecure to avoid common IP certificate mismatch in convenience path.
	to := TLSOptions{InsecureSkipVerify: true}
	return ExchangeTLSWithDialer(ctx, msg, server, d, to)
}
