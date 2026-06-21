// Package xray embeds xray-core as an in-process protocol engine and exposes
// it to sing-box as a custom outbound. The engine handles VLESS / XHTTP /
// Reality / Vision / Mux with xray's own implementations, so behaviour is
// byte-for-byte identical to servers configured with upstream xray.
package xray

import (
	"context"

	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"

	// Side-effect imports: register all xray protocol/transport features and
	// the JSON config loader. Without distro/all the VLESS/XHTTP/Reality stack
	// is not linked in; without main/json the "json" config format is unknown.
	_ "github.com/xtls/xray-core/main/distro/all"
	_ "github.com/xtls/xray-core/main/json"
)

// Engine wraps a started xray-core Instance configured with a single outbound
// (plus whatever routing the supplied fragment carries). All proxied traffic
// is dispatched through it via core.Dial / core.DialUDP.
type Engine struct {
	instance *core.Instance
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

// DialTCP opens a TCP connection to dest through the xray instance.
func (e *Engine) DialTCP(ctx context.Context, host string, port uint16) (xnet.Conn, error) {
	dest := xnet.TCPDestination(xnet.ParseAddress(host), xnet.Port(port))
	return core.Dial(ctx, e.instance, dest)
}

// DialUDP returns a PacketConn that dispatches UDP datagrams through the xray
// instance. Per-destination routing happens inside xray's dispatcher.
func (e *Engine) DialUDP(ctx context.Context) (xnet.PacketConn, error) {
	return core.DialUDP(ctx, e.instance)
}

// Close stops the underlying xray Instance.
func (e *Engine) Close() error {
	if e.instance == nil {
		return nil
	}
	return e.instance.Close()
}
