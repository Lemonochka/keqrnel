package xray

// Curated xray feature set for keqrnel, replacing main/distro/all.
//
// NOTE on what this does and does NOT trim: xray's JSON config parser
// (infra/conf, pulled in by main/json below) statically imports every proxy and
// transport package — vless/vmess/trojan/shadowsocks/hysteria/wireguard/tun/kcp/
// finalmask/… — and each registers its runtime handler via init(). So as long as
// we support JSON configs at all, all protocol/transport handlers are linked in
// regardless of this list; the protocol set cannot be slimmed without giving up
// the JSON loader. What this curation DOES drop vs distro/all: xray's own CLI
// commands, the gRPC commander/API + command apps, observatory, reverse, metrics,
// and the toml/yaml/external config loaders — none of which keqrnel uses.
import (
	// Core runtime apps (not pulled in by infra/conf — must be registered here).
	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/dns"
	_ "github.com/xtls/xray-core/app/log"
	_ "github.com/xtls/xray-core/app/policy"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	_ "github.com/xtls/xray-core/app/router"
	_ "github.com/xtls/xray-core/app/stats"

	// Tagged dialer, used by the DNS app / routing for outbound-tag dialing.
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"

	// JSON config format. This transitively links infra/conf and thus every
	// proxy/transport handler — that is xray's design, not an oversight.
	_ "github.com/xtls/xray-core/main/json"
)
