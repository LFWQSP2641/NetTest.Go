package transport

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DoHOptions controls minor DoH behaviors.
type DoHOptions struct {
	// Method: "POST" (default) or "GET" per RFC 8484.
	Method string
	// Extra headers to attach; optional.
	Headers http.Header
	// Timeouts used to configure http.Transport; zero values mean defaults.
	Timeouts Timeouts
}

const defaultDoHTimeout = 7 * time.Second

// NewHTTPTransportWithDialer builds an http.Transport using the provided Dialer and TLSOptions.
func NewHTTPTransportWithDialer(d Dialer, tlsOpt TLSOptions, t Timeouts) *http.Transport {
	tr := &http.Transport{
		// Inject custom dialer (SOCKS5/Direct)
		DialContext: d.DialContext,
		// TLS configuration derived from TLSOptions
		TLSClientConfig: tlsOpt.toTLSConfig(),
		// Timeouts
		TLSHandshakeTimeout:   durOrDefault(t.Handshake, 10*time.Second),
		ResponseHeaderTimeout: durOrDefault(t.Read, 0),
		IdleConnTimeout:       durOrDefault(t.Idle, 90*time.Second),
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	return tr
}

// ExchangeHTTPSWithRoundTripper performs a DoH request using the provided RoundTripper.
// dohURL like: https://dns.google/dns-query
func ExchangeHTTPSWithRoundTripper(ctx context.Context, msg *dns.Msg, dohURL string, rt http.RoundTripper, opt DoHOptions) (*dns.Msg, time.Duration, error) {
	if msg == nil {
		return nil, 0, errors.New("nil dns msg")
	}
	if dohURL == "" {
		return nil, 0, errors.New("empty doh url")
	}
	// Build client with the provided RT
	client := &http.Client{Transport: rt}

	// Pack DNS query
	payload, err := msg.Pack()
	if err != nil {
		return nil, 0, err
	}

	method := strings.ToUpper(opt.Method)
	if method == "" {
		method = http.MethodPost
	}

	var req *http.Request
	switch method {
	case http.MethodGet:
		// GET with base64url(dns) query param
		u, err := url.Parse(dohURL)
		if err != nil {
			return nil, 0, err
		}
		q := u.Query()
		q.Set("dns", base64.RawURLEncoding.EncodeToString(payload))
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("accept", "application/dns-message")
	default:
		// POST with application/dns-message body
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, dohURL, io.NopCloser(strings.NewReader(string(payload))))
		if err != nil {
			return nil, 0, err
		}
		req.Header.Set("content-type", "application/dns-message")
		req.Header.Set("accept", "application/dns-message")
		// Note: we used strings.NewReader(string(payload)) to avoid []byte.Copy in NewRequest; OK for small payloads
		// Alternatively use bytes.NewReader(payload)
	}

	// Extra headers
	for k, vs := range opt.Headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// read small body for error context
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, 0, errors.New("doh http status: " + resp.Status + ": " + string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	rtt := time.Since(start)

	out := new(dns.Msg)
	if err := out.Unpack(b); err != nil {
		return nil, 0, err
	}
	return out, rtt, nil
}

// ExchangeHTTPSWithDialer builds an http.Transport from the Dialer+TLSOptions, then performs DoH.
func ExchangeHTTPSWithDialer(ctx context.Context, msg *dns.Msg, dohURL string, d Dialer, tlsOpt TLSOptions, opt DoHOptions) (*dns.Msg, time.Duration, error) {
	// Ensure ServerName if not provided, derive from URL host
	if tlsOpt.ServerName == "" {
		if u, err := url.Parse(dohURL); err == nil {
			host := u.Hostname()
			if host != "" {
				tlsOpt.ServerName = host
			}
		}
	}
	tr := NewHTTPTransportWithDialer(d, tlsOpt, opt.Timeouts)
	return ExchangeHTTPSWithRoundTripper(ctx, msg, dohURL, tr, opt)
}

// ExchangeHTTPS convenience wrapper (direct connection, POST).
func ExchangeHTTPS(msg *dns.Msg, dohURL string) (*dns.Msg, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDoHTimeout)
	defer cancel()
	d := NewDirectDialer(DialOptions{Timeout: defaultDoHTimeout})
	// Default: secure verification using URL hostname as SNI
	var to TLSOptions
	return ExchangeHTTPSWithDialer(ctx, msg, dohURL, d, to, DoHOptions{Method: http.MethodPost})
}

// Helpers
func (t TLSOptions) toTLSConfig() *tls.Config {
	// implemented in options.go fields; mirror mapping here without duplicating type
	cfg := &tls.Config{
		ServerName:            t.ServerName,
		InsecureSkipVerify:    t.InsecureSkipVerify,
		RootCAs:               t.RootCAs,
		Certificates:          t.ClientCertificates,
		MinVersion:            t.MinVersion,
		MaxVersion:            t.MaxVersion,
		NextProtos:            t.NextProtos,
		VerifyPeerCertificate: t.VerifyPeerCertificate,
	}
	return cfg
}

func durOrDefault(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}
