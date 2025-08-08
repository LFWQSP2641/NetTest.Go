package dns

import (
	"crypto/tls"
	"encoding/json"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/txthinking/socks5"
)

func DnsRequest(server, qname, qtype, qclass string) string {
	m1 := buildDnsMassage(qname, qtype, qclass)

	dnsServer := server
	net := "udp"
	if strings.HasPrefix(dnsServer, "tcp://") {
		net = "tcp"
		dnsServer = strings.TrimPrefix(dnsServer, "tcp://")
	} else if strings.HasPrefix(dnsServer, "tls://") {
		net = "tcp-tls"
		dnsServer = strings.TrimPrefix(dnsServer, "tls://")
	}
	dnsServer = strings.TrimPrefix(dnsServer, "udp://")

	c := new(dns.Client)
	c.Net = net // Use UDP by default, can be changed to "tcp" if needed
	in, rtt, err := c.Exchange(m1, dnsServer)
	if err != nil {
		return getErrJsonResultString(err)
	}

	return getMassageResultString(in, rtt)
}

func DnsRequestOverSocks5(proxy, server, qname, qtype, qclass string) string {
	m1 := buildDnsMassage(qname, qtype, qclass)

	data, err := m1.Pack()
	if err != nil {
		return getErrJsonResultString(err)
	}

	dnsServer := server
	net := "udp"
	isTCP := false
	isTLS := false
	if strings.HasPrefix(dnsServer, "tcp://") {
		net = "tcp"
		isTCP = true
		dnsServer = strings.TrimPrefix(dnsServer, "tcp://")
	} else if strings.HasPrefix(dnsServer, "tls://") {
		net = "tcp"
		isTCP = true
		isTLS = true
		dnsServer = strings.TrimPrefix(dnsServer, "tls://")
	}
	dnsServer = strings.TrimPrefix(dnsServer, "udp://")

	proxyServer := strings.TrimPrefix(proxy, "socks5://")

	client, err := socks5.NewClient(proxyServer, "", "", 30, 30)
	if err != nil {
		return getErrJsonResultString(err)
	}

	conn, err := client.Dial(net, dnsServer)
	if err != nil {
		return getErrJsonResultString(err)
	}
	var tlsConn interface {
		Write([]byte) (int, error)
		Read([]byte) (int, error)
		SetDeadline(time.Time) error
		Close() error
	} = conn
	if isTLS {
		sni := dnsServer
		if idx := strings.Index(sni, ":"); idx > 0 {
			sni = sni[:idx]
		}
		tlsConfig := &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: false,
		}
		tlsConnRaw := tls.Client(conn, tlsConfig)
		err = tlsConnRaw.Handshake()
		if err != nil {
			conn.Close()
			return getErrJsonResultString(err)
		}
		tlsConn = tlsConnRaw
	}
	defer tlsConn.Close()

	if err := tlsConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return getErrJsonResultString(err)
	}
	start := time.Now()

	if isTCP {
		// DNS over TCP/TLS: 2 bytes length prefix
		msgLen := uint16(len(data))
		lenBuf := []byte{byte(msgLen >> 8), byte(msgLen & 0xff)}
		_, err = tlsConn.Write(lenBuf)
		if err == nil {
			_, err = tlsConn.Write(data)
		}
		if err != nil {
			return getErrJsonResultString(err)
		}
		// Read 2 bytes length
		lenBytes := make([]byte, 2)
		_, err = tlsConn.Read(lenBytes)
		if err != nil {
			return getErrJsonResultString(err)
		}
		respLen := int(lenBytes[0])<<8 | int(lenBytes[1])
		resp := make([]byte, respLen)
		n, err := tlsConn.Read(resp)
		if err != nil {
			return getErrJsonResultString(err)
		}
		in := &dns.Msg{}
		if err := in.Unpack(resp[:n]); err != nil {
			return getErrJsonResultString(err)
		}
		return getMassageResultString(in, time.Since(start))
	} else {
		// UDP
		_, err = tlsConn.Write(data)
		if err != nil {
			return getErrJsonResultString(err)
		}
		response := make([]byte, 1024)
		n, err := tlsConn.Read(response)
		if err != nil {
			return getErrJsonResultString(err)
		}
		in := &dns.Msg{}
		if err := in.Unpack(response[:n]); err != nil {
			return getErrJsonResultString(err)
		}
		return getMassageResultString(in, time.Since(start))
	}
}

func buildDnsMassage(qname, qtype, qclass string) *dns.Msg {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{Name: dns.Fqdn(qname), Qtype: dns.StringToType[qtype], Qclass: dns.StringToClass[qclass]}
	return m1
}

func getErrJsonResultString(err error) string {
	data := map[string]interface{}{
		"code":    -1,
		"message": err.Error(),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func getMassageResultString(m1 *dns.Msg, rtt time.Duration) string {
	data := map[string]interface{}{
		"rtt":    rtt,
		"answer": make([]map[string]interface{}, len(m1.Answer)),
		"flags": map[string]interface{}{
			"qr":     m1.Response,
			"opcode": m1.Opcode,
			"aa":     m1.Authoritative,
			"tc":     m1.Truncated,
			"rd":     m1.RecursionDesired,
			"ra":     m1.RecursionAvailable,
			"z":      m1.Zero,
			"ad":     m1.AuthenticatedData,
			"cd":     m1.CheckingDisabled,
			"rcode":  m1.Rcode,
		},
	}

	for i, ans := range m1.Answer {
		result := ""
		switch rr := ans.(type) {
		case *dns.A:
			result = rr.A.String()
		case *dns.AAAA:
			result = rr.AAAA.String()
		case *dns.CNAME:
			result = rr.Target
		default:
			result = rr.String()
		}
		data["answer"].([]map[string]interface{})[i] = map[string]interface{}{
			"name":   ans.Header().Name,
			"type":   dns.TypeToString[ans.Header().Rrtype],
			"class":  dns.ClassToString[ans.Header().Class],
			"ttl":    ans.Header().Ttl,
			"result": result,
			"data":   ans.String(),
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return getErrJsonResultString(err)
	}
	return string(jsonData)
}
