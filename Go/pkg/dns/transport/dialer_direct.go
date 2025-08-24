package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

type DirectDialer struct {
	d net.Dialer
}

func NewDirectDialer(opts DialOptions) *DirectDialer {
	d := net.Dialer{
		Timeout:   opts.Timeout,
		KeepAlive: opts.KeepAlive,
	}
	return &DirectDialer{d: d}
}

func (dd *DirectDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return dd.d.DialContext(ctx, network, address)
}

func (dd *DirectDialer) DialPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		// 解析远端地址
		raddr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("resolve udp addr: %w", err)
		}
		// 可选：基于 dd.d.LocalAddr 构造本地地址，这里用 nil 即自动选择
		pc, err := net.DialUDP(network, nil, raddr)
		if err != nil {
			return nil, fmt.Errorf("dial udp: %w", err)
		}
		// 将 ctx 的截止时间应用到底层连接（若有）
		if deadline, ok := ctx.Deadline(); ok {
			_ = pc.SetDeadline(deadline)
		}
		return pc, nil
	default:
		return nil, errors.New("DialPacket only supports udp/udp4/udp6 for DirectDialer")
	}
}

func DialContextWrapper(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dd := NewDirectDialer(DialOptions{Timeout: timeout})
	return dd.DialContext(ctx, network, address)
}

func DialPacketWrapper(ctx context.Context, network, address string, timeout time.Duration) (net.PacketConn, error) {
	dd := NewDirectDialer(DialOptions{Timeout: timeout})
	return dd.DialPacket(ctx, network, address)
}
