package dns

import (
	"context"
	"errors"
	"strings"
	"time"

	"nettest/pkg/dns/singdns"
	"nettest/pkg/dns/transport"

	"github.com/miekg/dns"
)

type DnsRequestType struct {
	id          string
	server      string
	net         string
	socks5Proxy string
	qname       string
	qtype       string
	qclass      string
}

type DnsResultType struct {
	id     string
	rtt    time.Duration
	answer *dns.Msg
}

func (d *DnsRequestType) Request() (*DnsResultType, error) {
	result := &DnsResultType{
		id: d.id,
	}
	if d.server == "" {
		return nil, errors.New("empty server")
	}

	// Build query message using existing helper
	msg := buildDnsMassage(d.qname, d.qtype, d.qclass)
	if msg == nil {
		return nil, errors.New("build dns message failed")
	}

	// Dialer selection (direct or socks5)
	// Expect socks5Proxy like "socks5://host:port" or "host:port"
	var dialer transport.Dialer
	timeout := 5 * time.Second
	switch d.net {
	case "tcp-tls", "https", "tls", "quic", "https3":
		timeout = 7 * time.Second
	}
	if p := strings.TrimSpace(d.socks5Proxy); p != "" {
		addr := strings.TrimPrefix(p, "socks5://")
		dialer = transport.NewSocks5Dialer(addr, "", "", transport.DialOptions{Timeout: timeout})
	} else {
		dialer = transport.NewDirectDialer(transport.DialOptions{Timeout: timeout})
	}

	// Address normalization handled by transport/singdns layer using d.server

	var (
		resp *dns.Msg
		rtt  time.Duration
		err  error
	)

	// Construct scheme-qualified address for sing-dns factory
	var serverAddr string
	switch d.net {
	case "udp":
		serverAddr = "udp://" + GetNetAddress(d.server)
	case "tcp":
		serverAddr = "tcp://" + GetNetAddress(d.server)
	case "tcp-tls", "tls":
		serverAddr = "tls://" + GetNetAddress(d.server)
	case "https":
		serverAddr = d.server // full DoH URL
	case "quic":
		if strings.HasPrefix(d.server, "quic://") || strings.HasPrefix(d.server, "doq://") {
			serverAddr = d.server
		} else {
			serverAddr = "quic://" + GetNetAddress(d.server)
		}
	case "https3":
		if strings.HasPrefix(d.server, "https3://") || strings.HasPrefix(d.server, "http3://") || strings.HasPrefix(d.server, "h3://") || strings.HasPrefix(d.server, "https://") {
			serverAddr = d.server
		} else {
			serverAddr = "https3://" + GetNetAddress(d.server)
		}
	default:
		serverAddr = d.server
	}

	switch d.net {
	case "udp":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// sing-dns path
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	case "tcp":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	case "tcp-tls", "tls":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	case "https":
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	case "quic":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	case "https3":
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()
		var pd transport.PacketDialer
		if v, ok := dialer.(transport.PacketDialer); ok {
			pd = v
		}
		sd := singdns.NewDialerAdapter(dialer, pd)
		t, e := singdns.CreateTransport(singdns.TransportOptions{Context: ctx, Dialer: sd, Address: serverAddr})
		if e != nil {
			err = e
			break
		}
		defer t.Close()
		start := time.Now()
		resp, err = t.Exchange(ctx, msg)
		rtt = time.Since(start)
	default:
		err = errors.New("unsupported net scheme: " + d.net)
	}

	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("empty response")
	}
	result.answer = resp
	result.rtt = rtt
	return result, nil
}
