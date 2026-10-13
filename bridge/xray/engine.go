// Package xray embeds xray-core as an in-process protocol engine and exposes
// it to sing-box as a custom outbound. The engine handles VLESS / XHTTP /
// Reality / Vision / Mux with xray's own implementations, so behaviour is
// byte-for-byte identical to servers configured with upstream xray.
package xray

import (
	"context"

	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	// xray feature registration (curated) lives in distro.go.
)

// Engine wraps a started xray-core Instance configured with a single outbound
// (plus whatever routing the supplied fragment carries). All proxied traffic
// is dispatched through it via core.Dial / core.DialUDP.
type Engine struct {
	instance *core.Instance
	sniffing *session.SniffingRequest
}

// NewEngine starts an xray Instance from a raw xray JSON config fragment.
// The fragment is the user's xray config as-is (outbounds + optional routing),
// which is what guarantees parity with their servers.
func NewEngine(configJSON []byte) (*Engine, error) {
	instance, err := core.StartInstance("json", configJSON)
	if err != nil {
		return nil, err
	}
	return &Engine{instance: instance}, nil
}

// SetSniffing makes xray sniff every connection handed to it, as its own
// inbounds do. sing-box passes the destination it was given, which under TUN
// is an IP even when it sniffed the domain itself: without this, xray's domain
// and geosite rules never see a domain there.
func (e *Engine) SetSniffing(request *session.SniffingRequest) {
	e.sniffing = request
}

// The dispatcher writes the sniffed protocol into the content, so each
// connection gets its own.
func (e *Engine) withSniffing(ctx context.Context) context.Context {
	if e.sniffing == nil || session.ContentFromContext(ctx) != nil {
		return ctx
	}
	return session.ContextWithContent(ctx, &session.Content{SniffingRequest: *e.sniffing})
}

// DialTCP opens a TCP connection to dest through the xray instance.
func (e *Engine) DialTCP(ctx context.Context, host string, port uint16) (xnet.Conn, error) {
	dest := xnet.TCPDestination(xnet.ParseAddress(host), xnet.Port(port))
	return core.Dial(e.withSniffing(ctx), e.instance, dest)
}

// DialUDP opens a connected UDP conn to dest through the xray instance. xray's
// core.Dial returns a net.Conn bound to the destination for UDP destinations,
// which is what sing-box's DialContext("udp", …) path (e.g. DNS-over-UDP via the
// proxy) expects.
func (e *Engine) DialUDP(ctx context.Context, host string, port uint16) (xnet.Conn, error) {
	dest := xnet.UDPDestination(xnet.ParseAddress(host), xnet.Port(port))
	return core.Dial(e.withSniffing(ctx), e.instance, dest)
}

// DialUDPPacket returns a PacketConn that dispatches UDP datagrams through the
// xray instance. Per-destination routing happens inside xray's dispatcher; used
// for the ListenPacket path (TUN UDP NAT sessions).
func (e *Engine) DialUDPPacket(ctx context.Context) (xnet.PacketConn, error) {
	return core.DialUDP(e.withSniffing(ctx), e.instance)
}

// Close stops the underlying xray Instance.
func (e *Engine) Close() error {
	if e.instance == nil {
		return nil
	}
	return e.instance.Close()
}
