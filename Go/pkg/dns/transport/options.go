package transport

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// Timeouts 统一控制各阶段超时；零值表示使用调用方或系统默认。
type Timeouts struct {
	// 建链超时（TCP/QUIC/HTTP 连接）
	Dial time.Duration
	// 握手超时（TLS/QUIC/HTTP2）
	Handshake time.Duration
	// 读/写超时（单次读写的截止时间）
	Read  time.Duration
	Write time.Duration
	// 空闲保活/连接池中的最大空闲时间（HTTP/2、HTTP/3等）
	Idle time.Duration
	// 整体操作截止时间（若上层未提供 context deadline，可用此作为兜底）
	Overall time.Duration
}

// TLSOptions 封装 TLS 配置；零值走系统默认根证书与版本。
type TLSOptions struct {
	// SNI/ServerName；为空则从目标主机推导。
	ServerName string
	// 是否跳过证书校验（仅用于测试/自签场景，生产应为 false）。
	InsecureSkipVerify bool

	// 根证书池；nil 时使用系统根证书。
	RootCAs *x509.CertPool
	// 客户端证书链；双向认证时填充。
	ClientCertificates []tls.Certificate

	// ALPN 协议，如 []string{"h2"} 或 {"h3", "h2"}。
	NextProtos []string

	// 协议版本范围；为 0 使用库默认。
	MinVersion uint16 // 例如 tls.VersionTLS12
	MaxVersion uint16 // 例如 tls.VersionTLS13

	// 证书固定（可选）：SPKI-SHA256 指纹集合（base64/raw 自行约定）
	SPKISHA256Pins [][]byte

	// 自定义验证回调；若非空，将在默认验证后调用（与 tls.Config.VerifyPeerCertificate 一致）。
	VerifyPeerCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error
}

type ProxyOptions struct {
	Type string
	Addr string
	User string
	Pass string
}
