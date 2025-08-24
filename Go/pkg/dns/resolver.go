package dns

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

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
	case "tcp-tls", "https":
		timeout = 7 * time.Second
	}
	if p := strings.TrimSpace(d.socks5Proxy); p != "" {
		addr := strings.TrimPrefix(p, "socks5://")
		dialer = transport.NewSocks5Dialer(addr, "", "", transport.DialOptions{Timeout: timeout})
	} else {
		dialer = transport.NewDirectDialer(transport.DialOptions{Timeout: timeout})
	}

	// Normalize address (strip scheme prefixes already handled elsewhere if needed)
	address := GetNetAddress(d.server)

	var (
		resp *dns.Msg
		rtt  time.Duration
		err  error
	)

	switch d.net {
	case "udp":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// Use PacketDialer; DirectDialer and Socks5Dialer both implement it
		if pd, ok := dialer.(transport.PacketDialer); ok {
			resp, rtt, err = transport.ExchangeUDPWithDialer(ctx, msg, address, pd)
		} else {
			err = errors.New("dialer does not support UDP packets")
		}
	case "tcp":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		resp, rtt, err = transport.ExchangeTCPWithDialer(ctx, msg, address, dialer)
	case "tcp-tls", "tls":
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// Derive SNI from host if present
		host, _, _ := net.SplitHostPort(address)
		tlsOpt := transport.TLSOptions{}
		if host != "" {
			if ip := net.ParseIP(host); ip != nil {
				// Connecting by IP: certificate name won't match; default to insecure for convenience.
				tlsOpt.InsecureSkipVerify = true
			} else {
				tlsOpt.ServerName = host
			}
		} else {
			// No host parsed, be permissive to avoid hard fail.
			tlsOpt.InsecureSkipVerify = true
		}
		resp, rtt, err = transport.ExchangeTLSWithDialer(ctx, msg, address, dialer, tlsOpt)
	case "https":
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()
		// For DoH we use the full URL in d.server; address has schemes stripped, so pass original
		tlsOpt := transport.TLSOptions{}
		resp, rtt, err = transport.ExchangeHTTPSWithDialer(ctx, msg, d.server, dialer, tlsOpt, transport.DoHOptions{Method: "POST"})
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
