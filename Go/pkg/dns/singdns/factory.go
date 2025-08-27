package singdns

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strings"

	"github.com/miekg/dns"
	upstreamdns "github.com/sagernet/sing-dns"
	_ "github.com/sagernet/sing-dns/quic"
	N "github.com/sagernet/sing/common/network"
)

type DomainStrategy int

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
	SNI          string
}

var transports map[string]TransportConstructor

func RegisterTransport(schemes []string, constructor TransportConstructor) {
	if transports == nil {
		transports = make(map[string]TransportConstructor)
	}
	for _, scheme := range schemes {
		transports[scheme] = constructor
	}
}

func CreateTransport(options TransportOptions) (Transport, error) {
	if transports == nil {
		return nil, errors.New("no transports registered")
	}
	constructor := transports[options.Address]
	if constructor == nil {
		if u, err := url.Parse(options.Address); err == nil && u != nil && u.Scheme != "" {
			constructor = transports[u.Scheme]
		}
	}
	if constructor == nil {
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
		m.SetEdns0(1232, true)
		opt = m.IsEdns0()
	}
	if opt == nil {
		return
	}
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

func init() {
	RegisterTransport([]string{"udp"}, func(opt TransportOptions) (Transport, error) {
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "udp", Dialer: opt.Dialer, Address: opt.Address})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "udp", upstream: u}, nil
	})
	RegisterTransport([]string{"tcp"}, func(opt TransportOptions) (Transport, error) {
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "tcp", Dialer: opt.Dialer, Address: opt.Address})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "tcp", upstream: u}, nil
	})
	RegisterTransport([]string{"tls"}, func(opt TransportOptions) (Transport, error) {
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "tls", Dialer: opt.Dialer, Address: opt.Address})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "tls", upstream: u}, nil
	})
	RegisterTransport([]string{"https"}, func(opt TransportOptions) (Transport, error) {
		u, err := upstreamdns.CreateTransport(upstreamdns.TransportOptions{Context: opt.Context, Name: "https", Dialer: opt.Dialer, Address: opt.Address})
		if err != nil {
			return nil, err
		}
		return singUpstreamTransport{name: "https", upstream: u}, nil
	})
	RegisterTransport([]string{"quic", "doq"}, func(opt TransportOptions) (Transport, error) {
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
	RegisterTransport([]string{"https3", "http3", "h3"}, func(opt TransportOptions) (Transport, error) {
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

// (no helpers required)

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
