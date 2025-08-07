package dns

import (
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
		data := map[string]interface{}{
			"error": err.Error(),
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		return string(jsonData)
	}

	dnsServer := server
	net := "udp"
	if strings.HasPrefix(dnsServer, "tcp://") {
		net = "tcp"
		dnsServer = strings.TrimPrefix(dnsServer, "tcp://")
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
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return getErrJsonResultString(err)
	}
	start := time.Now()
	_, err = conn.Write(data)
	if err != nil {
		return getErrJsonResultString(err)
	}

	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return getErrJsonResultString(err)
	}

	in := &dns.Msg{}
	if err := in.Unpack(response[:n]); err != nil {
		return getErrJsonResultString(err)
	}

	return getMassageResultString(in, time.Since(start))
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
		"code":  -1,
		"error": err.Error(),
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
