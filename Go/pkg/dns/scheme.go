package dns

import "strings"

func GetNetScheme(url string) string {
	net := "udp"
	if strings.HasPrefix(url, "tcp://") {
		net = "tcp"
	} else if strings.HasPrefix(url, "tls://") {
		net = "tcp-tls"
	} else if strings.HasPrefix(url, "https://") {
		net = "https"
	} else if strings.HasPrefix(url, "quic://") || strings.HasPrefix(url, "doq://") {
		net = "quic"
	} else if strings.HasPrefix(url, "https3://") || strings.HasPrefix(url, "http3://") || strings.HasPrefix(url, "h3://") {
		net = "https3"
	}
	return net
}

func GetNetAddress(url string) string {
	address := strings.TrimPrefix(url, "udp://")
	address = strings.TrimPrefix(address, "tcp://")
	address = strings.TrimPrefix(address, "tls://")
	address = strings.TrimPrefix(address, "https://")
	address = strings.TrimPrefix(address, "quic://")
	address = strings.TrimPrefix(address, "doq://")
	address = strings.TrimPrefix(address, "https3://")
	address = strings.TrimPrefix(address, "http3://")
	address = strings.TrimPrefix(address, "h3://")
	return address
}
