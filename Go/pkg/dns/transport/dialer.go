package transport

import (
	"context"
	"net"
	"time"
)

type DialOptions struct {
	Timeout   time.Duration
	KeepAlive time.Duration
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type PacketDialer interface {
	DialPacket(ctx context.Context, network, address string) (net.PacketConn, error)
}
