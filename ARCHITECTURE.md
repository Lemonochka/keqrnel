# Architecture

## Host and engine

The usual desktop setup runs xray for the protocol and sing-box for the TUN device,
with a loopback SOCKS hop between them: two processes, two configs, a port to
coordinate, and two chances to leave a TUN adapter behind on a crash. keqrnel keeps
both halves and puts them in one address space.

sing-box owns the process — TUN (`sing-tun`), routing, DNS, sniffing, `clash_api`,
and the lifecycle of everything below. xray-core is linked as a library and serves
one outbound type. Its protocols are executed by xray's own code, which is the reason
for embedding it rather than reimplementing anything.

```
TUN (sing-tun) ─► routing ─► outbounds:
                    ├─ direct / block
                    ├─ socks / http      (local upstream, e.g. wireproxy for AmneziaWG)
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

xray's `xnet.Conn` / `xnet.PacketConn` are aliases of the stdlib types, so the result
satisfies sing-box's `N.Dialer` with no shim.

The outbound implements `adapter.Lifecycle`: the engine starts in
`Start(StartStateStart)` and stops in `Close()`, so its lifetime stays paired even if
Box construction fails after the outbound was built
([bridge/xray/outbound.go](bridge/xray/outbound.go)).

## Configuration

One sing-box config; the bridge outbound carries a raw nested xray fragment that goes
to the engine as-is:

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

Binary size is the smaller reason. `include.*` also pulls in sing-box's own
`dns/transport/local`, whose Windows file calls a method the pinned `sing-tun`
renamed — curating the set is what makes the native Windows build compile.

sing-box always needs a transport named `local`: it is the default DNS fallback, used
for example to resolve the proxy server's own domain. keqrnel ships a small
cross-platform one over `net.Resolver` ([core/localdns/local.go](core/localdns/local.go)),
the same substitution libbox makes on Android.

## Packages

```
cmd/keqrnel/     CLI: config path, signals, stdin-EOF shutdown
service/         New / Start / Close, plus an optional platform hook
core/            registries + Box assembly
core/localdns/   the "local" DNS transport
bridge/xray/     engine.go (embedded instance), outbound.go (bridge),
                 distro.go (curated xray feature set)
```

Hazards and known limits: [ERROR-ANALYSIS.md](ERROR-ANALYSIS.md).
