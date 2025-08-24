package dns

import (
	"encoding/json"
	"errors"
	"time"

	utils "nettest/pkg/utils"

	"github.com/miekg/dns"
)

func DnsRequest(server, qname, qtype, qclass string) string {
	// Thin facade over resolver.Request
	netScheme := GetNetScheme(server)
	req := &DnsRequestType{
		id:     "",
		server: server,
		net:    netScheme,
		qname:  qname,
		qtype:  qtype,
		qclass: qclass,
	}
	res, err := req.Request()
	if err != nil {
		return utils.BuildErrJSON(err)
	}
	return getMassageResultString(res.answer, res.rtt)
}

func DnsRequestOverSocks5(proxy, server, qname, qtype, qclass string) string {
	netScheme := GetNetScheme(server)
	req := &DnsRequestType{
		id:          "",
		server:      server,
		net:         netScheme,
		socks5Proxy: proxy,
		qname:       qname,
		qtype:       qtype,
		qclass:      qclass,
	}
	res, err := req.Request()
	if err != nil {
		return utils.BuildErrJSON(err)
	}
	return getMassageResultString(res.answer, res.rtt)
}

func buildDnsMassage(qname, qtype, qclass string) *dns.Msg {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{Name: dns.Fqdn(qname), Qtype: dns.StringToType[qtype], Qclass: dns.StringToClass[qclass]}
	return m1
}

func getMassageResultString(m1 *dns.Msg, rtt time.Duration) string {
	if m1 == nil {
		return utils.BuildErrJSON(errors.New("nil dns message"))
	}
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
		return utils.BuildErrJSON(err)
	}
	return string(jsonData)
}
