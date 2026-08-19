# Architecture

## Host and engine

sing-box owns the process: TUN (`sing-tun`), routing, DNS, sniffing, `clash_api`,
and the lifecycle of everything below it. xray-core is linked as a library and serves
one outbound type; its protocols are executed by xray's own code.

The alternative is xray for the protocol and sing-box for the TUN with a loopback
SOCKS hop between them, which needs two processes and a port they agree on.

```
TUN (sing-tun) ─► routing ─► outbounds:
                    ├─ direct / block
                    ├─ socks / http      (local upstream)
                    └─ xray  ──────────► embedded xray.Instance
```

## The bridge

xray exports `core.Dial(ctx, instance, dest) (net.Conn, error)`, which dials through
that instance's own routing and outbounds. The bridge outbound calls it:

| sing-box asks for | bridge calls | used by |
|---|---|---|
| `DialContext("tcp")` | `core.Dial`, TCP destination | all TCP |
| `DialContext("udp")` | `core.Dial`, UDP destination | DNS-over-UDP through the proxy |
| `ListenPacket` | `core.DialUDP` | UDP sessions from the TUN's NAT |

xray's `xnet.Conn` and `xnet.PacketConn` are aliases of the stdlib types, so the
result satisfies sing-box's `N.Dialer` directly.

The outbound implements `adapter.Lifecycle`: the engine starts in
`Start(StartStateStart)` and stops in `Close()`, so it cannot outlive a Box whose
construction failed after the outbound was built
([bridge/xray/outbound.go](bridge/xray/outbound.go)).

## Configuration

One sing-box config; the bridge outbound carries a nested xray fragment that goes to
the engine as-is:

```json
{ "type": "xray", "tag": "proxy", "xray": { "outbounds": [ ... ] } }
```

An outbound copied out of a working xray config keeps working. Routing, DNS and TUN
stay on the sing-box side; the fragment can carry its own `routing` block if you want
xray to route internally as well.

## Curated registries

[core/registry.go](core/registry.go) builds the registries by hand instead of using
`include.*`:

| Registry | Registered | Left out |
|---|---|---|
| inbound | tun, mixed, socks, http | everything server-side |
| outbound | direct, block, selector, urltest, socks, http, **xray** | tor, naive, ssh, and the protocols xray now serves |
| DNS transport | udp, tcp, tls, https, hosts, fakeip, **local** | dhcp, resolved, other platform glue |
| endpoint | — | WireGuard/AmneziaWG stay a separate core |

`include.*` also pulls in sing-box's `dns/transport/local`, whose Windows file calls
a method the pinned `sing-tun` renamed, so it does not compile there. A transport
named `local` is nevertheless mandatory — sing-box creates a default DNS fallback of
that type at startup, and without one it fails with
`transport type not found: local` — so keqrnel registers a small cross-platform one
over `net.Resolver` ([core/localdns/local.go](core/localdns/local.go)).

## Notes

- `DialContext` handles TCP and connected UDP. Unconnected UDP is rejected and goes
  through `ListenPacket`, matching sing-box's TUN NAT model.
- Each UDP flow gets its own `ListenPacket` and therefore its own `dispatcherConn`;
  `WriteTo` carries the destination per packet.
- A UDP goroutine can outlive its session by up to a minute: xray's `dispatcherConn`
  reclaims the entry on an inactivity timer rather than on `Close()`.
- The xray fragment is passed through untouched, so a server that rejects the
  connection is a question about the fragment, not about keqrnel.

## Packages

```
cmd/keqrnel/     CLI: config path, signals, stdin-EOF shutdown
service/         New / Start / Close
core/            registries + Box assembly
core/localdns/   the "local" DNS transport
bridge/xray/     engine.go (embedded instance), outbound.go (bridge),
                 distro.go (curated xray feature set)
```
