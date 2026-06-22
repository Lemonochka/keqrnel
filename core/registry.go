// Package core assembles the keqrnel host: a sing-box Box wired with a curated
// set of registries (only what keqrnel actually uses, no legacy) plus the
// embedded-xray bridge outbound ("xray").
//
// We deliberately do NOT use sing-box's include.* helpers: they register every
// protocol/transport sing-box ships (tor, naive, ssh, dhcp-dns, the broken
// windows local-resolver, ...). Curating the registries keeps the binary lean,
// drops legacy we'll never use, and avoids upstream platform-specific breakage.
package core

import (
	"context"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	boxservice "github.com/sagernet/sing-box/adapter/service"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport"
	"github.com/sagernet/sing-box/dns/transport/fakeip"
	"github.com/sagernet/sing-box/dns/transport/hosts"
	"github.com/sagernet/sing-box/experimental/deprecated"
	// clash_api server (traffic counters over HTTP); enabled per-config.
	_ "github.com/sagernet/sing-box/experimental/clashapi"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/protocol/block"
	"github.com/sagernet/sing-box/protocol/direct"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing-box/protocol/http"
	"github.com/sagernet/sing-box/protocol/mixed"
	"github.com/sagernet/sing-box/protocol/socks"
	"github.com/sagernet/sing-box/protocol/tun"
	"github.com/sagernet/sing/service"

	xraybridge "github.com/keqdroid/keqrnel/bridge/xray"
	"github.com/keqdroid/keqrnel/core/localdns"
)

// Context returns a context carrying keqrnel's curated registries plus the xray
// bridge outbound. Use it both to decode the config (so "xray" outbound options
// unmarshal) and to construct the Box.
func Context(ctx context.Context) context.Context {
	// A deprecated-feature manager is expected in context while options are
	// decoded; mirror sing-box's own CLI setup.
	ctx = service.ContextWith[deprecated.Manager](ctx, deprecated.NewStderrManager(log.StdLogger()))

	return box.Context(
		ctx,
		inboundRegistry(),
		outboundRegistry(),
		endpoint.NewRegistry(), // no endpoints: AmneziaWG/WireGuard stay a separate core
		dnsTransportRegistry(),
		boxservice.NewRegistry(), // no extra services
	)
}

// inboundRegistry: TUN (the whole point) plus local proxy listeners.
func inboundRegistry() *inbound.Registry {
	registry := inbound.NewRegistry()
	tun.RegisterInbound(registry)
	mixed.RegisterInbound(registry)
	socks.RegisterInbound(registry)
	http.RegisterInbound(registry)
	return registry
}

// outboundRegistry: direct/block for routing, selector/urltest for groups, and
// the embedded-xray bridge that serves VLESS/XHTTP/Reality/Vision/Mux + Hysteria2.
func outboundRegistry() *outbound.Registry {
	registry := outbound.NewRegistry()
	direct.RegisterOutbound(registry)
	block.RegisterOutbound(registry)
	group.RegisterSelector(registry)
	group.RegisterURLTest(registry)
	// socks/http outbounds: keqrnel doubles as the generic sing-box TUN engine
	// (e.g. AmneziaWG: wireproxy SOCKS -> keqrnel TUN), so it must dial a local
	// socks/http upstream too, not only the embedded-xray bridge.
	socks.RegisterOutbound(registry)
	http.RegisterOutbound(registry)
	xraybridge.Register(registry)
	return registry
}

// dnsTransportRegistry: modern remote DNS only (udp/tcp/tls/https) + hosts +
// fakeip. No system "local"/"resolved"/dhcp transports (legacy + platform glue).
func dnsTransportRegistry() *dns.TransportRegistry {
	registry := dns.NewTransportRegistry()
	transport.RegisterUDP(registry)
	transport.RegisterTCP(registry)
	transport.RegisterTLS(registry)
	transport.RegisterHTTPS(registry)
	hosts.RegisterTransport(registry)
	fakeip.RegisterTransport(registry)
	localdns.RegisterTransport(registry) // keqrnel's own "local" (default DNS fallback)
	return registry
}
