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
	}
	return net
}

func GetNetAddress(url string) string {
	address := strings.TrimPrefix(url, "udp://")
	address = strings.TrimPrefix(address, "tcp://")
	address = strings.TrimPrefix(address, "tls://")
	address = strings.TrimPrefix(address, "https://")
	return address
}
