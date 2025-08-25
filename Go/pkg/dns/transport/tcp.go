package transport

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/miekg/dns"
)

const defaultTCPTimeout = 5 * time.Second

// ExchangeTCPWithDialer sends a DNS message over TCP using the provided Dialer.
// It implements DNS over TCP framing: 2-byte length prefix followed by the payload.
// - server: "host:port" (e.g., "8.8.8.8:53")
// - d: a Dialer, e.g., DirectDialer or Socks5Dialer
func ExchangeTCPWithDialer(ctx context.Context, msg *dns.Msg, server string, d Dialer) (*dns.Msg, time.Duration, error) {
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

	conn, err := d.DialContext(ctx, "tcp", server)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	// Per-stage deadlines (write then read) to avoid overly broad deadlines
	var writeDeadline, readDeadline time.Time
	if dl, ok := ctx.Deadline(); ok {
		writeDeadline, readDeadline = dl, dl
	}

	start := time.Now()

	// Build single contiguous buffer: [2-byte len][payload]
	frame := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(payload)))
	copy(frame[2:], payload)
	if !writeDeadline.IsZero() {
		_ = conn.SetWriteDeadline(writeDeadline)
	}
	if err := writeFull(conn, frame); err != nil {
		return nil, 0, err
	}

	// Read length-prefixed response
	if !readDeadline.IsZero() {
		_ = conn.SetReadDeadline(readDeadline)
	}
	var hp [2]byte
	if _, err := io.ReadFull(conn, hp[:]); err != nil {
		return nil, 0, err
	}
	n := int(binary.BigEndian.Uint16(hp[:]))
	if n <= 0 {
		return nil, 0, errors.New("empty dns tcp response")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, 0, err
	}

	rtt := time.Since(start)

	resp := new(dns.Msg)
	if err := resp.Unpack(buf); err != nil {
		return nil, 0, err
	}
	return resp, rtt, nil
}

// ExchangeTCP is a convenience wrapper for direct TCP without a proxy.
// It uses a DirectDialer with a default timeout and a context with the same timeout.
func ExchangeTCP(msg *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTCPTimeout)
	defer cancel()
	d := NewDirectDialer(DialOptions{Timeout: defaultTCPTimeout})
	return ExchangeTCPWithDialer(ctx, msg, server, d)
}

// writeFull writes the entire buffer or returns an error.
func writeFull(w io.Writer, b []byte) error {
	// Try Write + Handle short writes
	for len(b) > 0 {
		n, err := w.Write(b)
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return err
		}
		if err != nil {
			return err
		}
		b = b[n:]
	}
	return nil
}
