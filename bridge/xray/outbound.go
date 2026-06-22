package xray

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

// TypeName is the outbound type used in keqrnel config: { "type": "xray", ... }.
const TypeName = "xray"

// Options is the config schema for the xray bridge outbound. Config carries the
// user's raw xray JSON fragment (outbounds + optional routing), used verbatim.
type Options struct {
	Config json.RawMessage `json:"xray"`
}

// Register wires the xray bridge outbound into a sing-box outbound registry.
// Call once during registry construction, before building the Box.
func Register(registry *outbound.Registry) {
	outbound.Register[Options](registry, TypeName, New)
}

// Outbound is a sing-box outbound backed by an embedded xray engine. The engine
// is started in Start (not New) so its lifetime is paired with Close even if
// Box construction fails afterwards, and so config errors surface at start time.
type Outbound struct {
	outbound.Adapter
	logger log.ContextLogger
	config json.RawMessage

	mu     sync.RWMutex
	engine *Engine
}

var (
	_ adapter.Outbound  = (*Outbound)(nil)
	_ adapter.Lifecycle = (*Outbound)(nil)
	_ N.Dialer          = (*Outbound)(nil)
)

// New validates options and constructs the outbound; the embedded xray engine
// is started later, in Start.
func New(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options Options) (adapter.Outbound, error) {
	if len(options.Config) == 0 {
		return nil, E.New("missing \"xray\" config fragment")
	}
	return &Outbound{
		Adapter: outbound.NewAdapter(TypeName, tag, []string{N.NetworkTCP, N.NetworkUDP}, nil),
		logger:  logger,
		config:  options.Config,
	}, nil
}

// Start boots the embedded xray engine when the Box reaches the start stage.
func (o *Outbound) Start(stage adapter.StartStage) error {
	if stage != adapter.StartStateStart {
		return nil
	}
	engine, err := NewEngine(o.config)
	if err != nil {
		return E.Cause(err, "start embedded xray engine")
	}
	o.mu.Lock()
	o.engine = engine
	o.mu.Unlock()
	return nil
}

func (o *Outbound) currentEngine() (*Engine, error) {
	o.mu.RLock()
	engine := o.engine
	o.mu.RUnlock()
	if engine == nil {
		return nil, E.New("xray engine not started")
	}
	return engine, nil
}

// DialContext dispatches a connection through the xray engine. TCP returns a
// stream conn; UDP returns a connected packet conn (used e.g. for DNS-over-UDP
// through the proxy, which sing-box dials via DialContext rather than ListenPacket).
func (o *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	engine, err := o.currentEngine()
	if err != nil {
		return nil, err
	}
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		return engine.DialTCP(ctx, host(destination), destination.Port)
	case N.NetworkUDP:
		return engine.DialUDP(ctx, host(destination), destination.Port)
	default:
		return nil, E.New("xray outbound: unsupported network: ", network)
	}
}

// ListenPacket dispatches UDP datagrams through the xray engine. The returned
// PacketConn carries each datagram's destination per-packet, so a single conn
// serves the whole UDP session that sing-box's NAT routed here.
func (o *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	engine, err := o.currentEngine()
	if err != nil {
		return nil, err
	}
	return engine.DialUDPPacket(ctx)
}

// Close stops the embedded xray engine.
func (o *Outbound) Close() error {
	o.mu.Lock()
	engine := o.engine
	o.engine = nil
	o.mu.Unlock()
	if engine == nil {
		return nil
	}
	return engine.Close()
}

// host extracts a dialable host (domain or IP literal) from a sing socksaddr.
func host(destination M.Socksaddr) string {
	if destination.IsFqdn() {
		return destination.Fqdn
	}
	return destination.Addr.String()
}
