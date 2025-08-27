package singdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"

	"nettest/pkg/dns/transport"

	"github.com/miekg/dns"
	upstreamdns "github.com/sagernet/sing-dns"
	_ "github.com/sagernet/sing-dns/quic"
	N "github.com/sagernet/sing/common/network"
)

// DomainStrategy is a placeholder for strategy selection; adjust as needed.
type DomainStrategy int

// Transport is a minimal interface to unify different DNS transports.
type Transport interface {
	Name() string
	Start() error
	Reset()
	Close() error
	Raw() bool
	Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error)
	Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error)
}

type TransportConstructor func(TransportOptions) (Transport, error)

type TransportOptions struct {
	Context      context.Context
	Name         string
	Dialer       N.Dialer
	Address      string
	ClientSubnet netip.Prefix
}

var transports map[string]TransportConstructor

// RegisterTransport registers a constructor for given schemes (or full addresses).
func RegisterTransport(schemes []string, constructor TransportConstructor) {
	if transports == nil {
		transports = make(map[string]TransportConstructor)
	}
	for _, scheme := range schemes {
		transports[scheme] = constructor
	}
}

// CreateTransport locates a constructor by full address key, or by URL scheme.
// e.g. "udp://1.1.1.1:53" → lookup key "udp"; "https://dns.google/dns-query" → key "https".
func CreateTransport(options TransportOptions) (Transport, error) {
	if transports == nil {
		return nil, errors.New("no transports registered")
	}
	constructor := transports[options.Address]
	if constructor == nil {
		// Try parse as URL to get scheme
		if u, err := url.Parse(options.Address); err == nil && u != nil && u.Scheme != "" {
			constructor = transports[u.Scheme]
		}
	}
	if constructor == nil {
		// Fallback: simple prefix probe
		for k := range transports {
			if len(k) > 0 && len(options.Address) >= len(k)+3 && options.Address[:len(k)+3] == k+"://" {
				constructor = transports[k]
				break
			}
		}
	}
	if constructor == nil {
		return nil, fmt.Errorf("unknown DNS server format: %s", options.Address)
	}
	// Bind name into context for downstream if needed (no-op here)
	if options.Context == nil {
		options.Context = context.Background()
	}
	t, err := constructor(options)
	if err != nil {
		return nil, err
	}
	if options.ClientSubnet.IsValid() {
		t = &edns0SubnetTransportWrapper{inner: t, subnet: options.ClientSubnet}
	}
	return t, nil
}

// edns0SubnetTransportWrapper injects ECS before calling inner transport.
type edns0SubnetTransportWrapper struct {
	inner  Transport
	subnet netip.Prefix
}

func (w *edns0SubnetTransportWrapper) Name() string { return w.inner.Name() }
func (w *edns0SubnetTransportWrapper) Start() error { return w.inner.Start() }
func (w *edns0SubnetTransportWrapper) Reset()       { w.inner.Reset() }
func (w *edns0SubnetTransportWrapper) Close() error { return w.inner.Close() }
func (w *edns0SubnetTransportWrapper) Raw() bool    { return w.inner.Raw() }
func (w *edns0SubnetTransportWrapper) Lookup(ctx context.Context, d string, s DomainStrategy) ([]netip.Addr, error) {
	return w.inner.Lookup(ctx, d, s)
}

func (w *edns0SubnetTransportWrapper) Exchange(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	if m != nil && w.subnet.IsValid() {
		ensureECS(m, w.subnet)
	}
	return w.inner.Exchange(ctx, m)
}

func ensureECS(m *dns.Msg, p netip.Prefix) {
	opt := m.IsEdns0()
	if opt == nil {
		m.SetEdns0(1232, true) // typical UDP size
		opt = m.IsEdns0()
	}
	if opt == nil {
		return
	}
	// Remove existing ECS if any
	var rest []dns.EDNS0
	for _, o := range opt.Option {
		if o.Option() != dns.EDNS0SUBNET {
			rest = append(rest, o)
		}
	}
	opt.Option = rest

	var family uint16
	var addr []byte
	ip := p.Addr()
	if ip.Is4() {
		family = 1
		addr = ip.AsSlice()[:4]
	} else {
		family = 2
		addr = ip.AsSlice()[:16]
	}
	ecs := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        family,
		SourceNetmask: uint8(p.Bits()),
		SourceScope:   0,
		Address:       addr,
	}
	opt.Option = append(opt.Option, ecs)
}

// Built-in constructors using our transport layer, keyed by scheme.
func init() {
	// UDP
	RegisterTransport([]string{"udp"}, func(opt TransportOptions) (Transport, error) {
		return simpleExchangeTransport{
			name:    "udp",
			dialer:  opt.Dialer,
			address: opt.Address,
			exch: func(ctx context.Context, m *dns.Msg, d N.Dialer, addr string) (*dns.Msg, error) {
				// Use our PacketDialer if available
				if da, ok := d.(*dialerAdapter); ok {
					if pd := da.OurPacketDialer(); pd != nil {
						r, _, err := transport.ExchangeUDPWithDialer(ctx, m, stripScheme(addr), pd)
						return r, err
					}
				}
				// QUIC (DoQ) via sing-dns upstream registry
				RegisterTransport([]string{"quic", "doq"}, func(opt TransportOptions) (Transport, error) {
					if opt.Dialer == nil {
						return nil, errors.New("nil dialer for quic")
					}
					a := opt.Address
					if strings.HasPrefix(a, "doq://") {
						a = "quic://" + a[len("doq://"):]
					}
					u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{
						Context: opt.Context,
						Name:    "quic",
						Dialer:  opt.Dialer,
						Address: a,
					})
					if err != nil {
						return nil, err
					}
					return singUpstreamTransport{name: "quic", upstream: u}, nil
				})

				// HTTP/3 (DoH3) via sing-dns upstream registry
				RegisterTransport([]string{"https3", "http3", "h3"}, func(opt TransportOptions) (Transport, error) {
					if opt.Dialer == nil {
						return nil, errors.New("nil dialer for http3")
					}
					a := opt.Address
					if u, err := url.Parse(a); err == nil && u != nil {
						if u.Scheme == "https3" || u.Scheme == "http3" {
							u.Scheme = "h3"
							a = u.String()
						}
					}
					u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{
						Context: opt.Context,
						Name:    "h3",
						Dialer:  opt.Dialer,
						Address: a,
					})
					if err != nil {
						return nil, err
					}
					return singUpstreamTransport{name: "https3", upstream: u}, nil
				})
				return nil, transport.ErrSocks5UDPUnsupported
			},
		}, nil
	})
	// TCP
	RegisterTransport([]string{"tcp"}, func(opt TransportOptions) (Transport, error) {
		return simpleExchangeTransport{
			name:    "tcp",
			dialer:  opt.Dialer,
			address: opt.Address,
			exch: func(ctx context.Context, m *dns.Msg, d N.Dialer, addr string) (*dns.Msg, error) {
				if da, ok := d.(*dialerAdapter); ok {
					r, _, err := transport.ExchangeTCPWithDialer(ctx, m, stripScheme(addr), da.OurDialer())
					return r, err
				}
				return nil, errors.New("incompatible dialer adapter")
			},
		}, nil
	})
	// TLS (DoT)
	RegisterTransport([]string{"tls"}, func(opt TransportOptions) (Transport, error) {
		return simpleExchangeTransport{
			name:    "tls",
			dialer:  opt.Dialer,
			address: opt.Address,
			exch: func(ctx context.Context, m *dns.Msg, d N.Dialer, addr string) (*dns.Msg, error) {
				hostPort := stripScheme(addr)
				// Derive SNI/insecure from host
				tlsOpt := transport.TLSOptions{}
				host := hostOnly(hostPort)
				if host != "" {
					if ip := net.ParseIP(host); ip != nil {
						tlsOpt.InsecureSkipVerify = true
					} else {
						tlsOpt.ServerName = host
					}
				} else {
					tlsOpt.InsecureSkipVerify = true
				}
				if da, ok := d.(*dialerAdapter); ok {
					r, _, err := transport.ExchangeTLSWithDialer(ctx, m, hostPort, da.OurDialer(), tlsOpt)
					return r, err
				}
				return nil, errors.New("incompatible dialer adapter")
			},
		}, nil
	})
	// HTTPS (DoH)
	RegisterTransport([]string{"https"}, func(opt TransportOptions) (Transport, error) {
		return simpleExchangeTransport{
			name:    "https",
			dialer:  opt.Dialer,
			address: opt.Address,
			exch: func(ctx context.Context, m *dns.Msg, d N.Dialer, addr string) (*dns.Msg, error) {
				if da, ok := d.(*dialerAdapter); ok {
					// Build TLSOptions: SNI for hostname, insecure for IP
					tlsOpt := transport.TLSOptions{}
					if u, err := url.Parse(addr); err == nil && u != nil {
						host := u.Hostname()
						if host != "" {
							if ip := net.ParseIP(host); ip != nil {
								tlsOpt.InsecureSkipVerify = true
							} else {
								tlsOpt.ServerName = host
							}
						}
					}
					r, _, err := transport.ExchangeHTTPSWithDialer(ctx, m, addr, da.OurDialer(), tlsOpt, transport.DoHOptions{Method: "POST"})
					return r, err
				}
				return nil, errors.New("incompatible dialer adapter")
			},
		}, nil
	})

	// QUIC (DoQ) via sing-dns upstream
	RegisterTransport([]string{"quic", "doq"}, func(opt TransportOptions) (Transport, error) {
		if opt.Dialer == nil {
			return nil, errors.New("nil dialer for quic")
		}
		a := opt.Address
		if strings.HasPrefix(a, "doq://") {
			a = "quic://" + a[len("doq://"):]
		}
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "quic", Dialer: opt.Dialer, Address: a})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "quic", upstream: u}, nil
	})

	// HTTP/3 (DoH3) via sing-dns upstream
	RegisterTransport([]string{"https3", "http3", "h3"}, func(opt TransportOptions) (Transport, error) {
		if opt.Dialer == nil {
			return nil, errors.New("nil dialer for http3")
		}
		a := opt.Address
		if u, err := url.Parse(a); err == nil && u != nil {
			if u.Scheme == "https3" || u.Scheme == "http3" {
				u.Scheme = "h3"
				a = u.String()
			}
		}
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "h3", Dialer: opt.Dialer, Address: a})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "https3", upstream: u}, nil
	})
}

// simpleExchangeTransport is a thin wrapper to reuse our transport exchangers.
type simpleExchangeTransport struct {
	name    string
	dialer  N.Dialer
	address string
	exch    func(ctx context.Context, m *dns.Msg, d N.Dialer, addr string) (*dns.Msg, error)
}

func (t simpleExchangeTransport) Name() string { return t.name }
func (t simpleExchangeTransport) Start() error { return nil }
func (t simpleExchangeTransport) Reset()       {}
func (t simpleExchangeTransport) Close() error { return nil }
func (t simpleExchangeTransport) Raw() bool    { return false }
func (t simpleExchangeTransport) Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error) {
	return nil, errors.New("Lookup not implemented in simpleExchangeTransport")
}
func (t simpleExchangeTransport) Exchange(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	return t.exch(ctx, m, t.dialer, t.address)
}

func stripScheme(addr string) string {
	// strip known schemes to get host:port
	for _, s := range []string{"udp://", "tcp://", "tls://"} {
		if len(addr) > len(s) && addr[:len(s)] == s {
			return addr[len(s):]
		}
	}
	return addr
}

// hostOnly extracts host from host:port or returns input if no colon.
func hostOnly(hostPort string) string {
	if i := strings.LastIndex(hostPort, ":"); i > 0 {
		return hostPort[:i]
	}
	return hostPort
}

// singUpstreamTransport wraps a sing-dns upstream transport to our Transport.
type singUpstreamTransport struct {
	name     string
	upstream interface {
		Exchange(ctx context.Context, m *dns.Msg) (*dns.Msg, error)
		Close() error
		Start() error
		Raw() bool
		Reset()
	}
}

func (t singUpstreamTransport) Name() string { return t.name }
func (t singUpstreamTransport) Start() error { return t.upstream.Start() }
func (t singUpstreamTransport) Reset()       { t.upstream.Reset() }
func (t singUpstreamTransport) Close() error { return t.upstream.Close() }
func (t singUpstreamTransport) Raw() bool    { return t.upstream.Raw() }
func (t singUpstreamTransport) Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error) {
	return nil, errors.New("Lookup not implemented in singUpstreamTransport")
}
func (t singUpstreamTransport) Exchange(ctx context.Context, m *dns.Msg) (*dns.Msg, error) {
	return t.upstream.Exchange(ctx, m)
}
