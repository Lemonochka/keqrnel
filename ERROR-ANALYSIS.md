# Hazards, trade-offs and limits

What auditing keqrnel against the sing-box v1.13.13 and xray-core v1.260327.0 sources
turned up: the failure modes embedding one core inside another produces, and what is
left alone on purpose.

## Lifecycle

**The xray instance could leak on a partial `box.New` failure.** The engine used to
start in `New()`, while the Box was still under construction; if a later outbound
failed to build, our `Close()` was never called. The outbound now implements
`adapter.Lifecycle` — engine starts in `Start(StartStateStart)`, stops in `Close()` —
so the pairing holds wherever construction fails.

`Close()` is reached on both paths: `outbound.Manager.Close()` closes every outbound
through `io.Closer`, and a reload that replaces or drops one calls
`common.Close(existsOutbound)`.

**Engine race on shutdown.** The `engine` field is under a `sync.RWMutex` and nilled
on close; `currentEngine()` returns an error instead of dereferencing nil when a dial
races a close.

**A UDP goroutine can linger up to ~60 s after its session ends.** xray's
`dispatcherConn` reclaims the `connEntry` on a one-minute inactivity timer, not on
`Close()`. Accepted: the count stays bounded by active sessions.

## Network

**UDP is routed per destination.** Every UDP flow out of the TUN gets its own
`ListenPacket` and therefore its own `dispatcherConn`, and `WriteTo` carries the
destination in each packet. The embedded instance is effectively a single outbound
(routing lives on the sing-box side), so multiplexing onto one link is unambiguous.

**`DialContext` handles TCP and connected UDP only.** Unconnected UDP is rejected
with an explicit error and goes through `ListenPacket`, matching sing-box's TUN NAT
model.

`host()` yields an FQDN or a string IP, and `xnet.ParseAddress` tells domain, IPv4
and IPv6 apart correctly.

## Platform

**sing-box v1.13.13 does not build on Windows.** Its pinned `sing-tun` renamed
`MyInterface()` to `MyInterfaces()`, while `dns/transport/local/resolv_windows.go`
still calls the old name. The file is Windows-only, so Linux is unaffected. keqrnel
avoids `include.*`, which would pull that package in, and curates registries in
[core/registry.go](core/registry.go).

**A `local` DNS transport is mandatory.** sing-box creates a default DNS fallback of
that type at startup; without one registered it fails with
`default DNS server fallback: transport type not found: local`. Ours is backed by
`net.Resolver` ([core/localdns/local.go](core/localdns/local.go)).

Native Windows/amd64 and cross linux/amd64, linux/arm64, darwin/arm64 all build clean.

## Config

The `xray` field is taken 1:1 and xray executes the protocols, so what reaches the
server matches upstream xray byte for byte. Malformed configs and an empty `xray`
fragment return errors rather than panicking (`TestInvalidConfigReports`, plus the
check in `New`).

**Interop with live servers is not covered by the tests.** They prove the engine
moves traffic; whether a given Reality / XHTTP / Hysteria2 server is happy depends on
the fragment you feed in, which keqrnel passes through untouched. That translation is
the caller's job and the first place to look when one specific server misbehaves.

## Tests

| Package | Covers |
|---|---|
| `bridge/xray` | TCP and UDP round-trips through a real `core.Dial` |
| `core` | the example config decodes through the curated registries |
| `service` | full New / Start / Close with route rules; error on a broken config |

Beyond that, the binary has been smoke-tested end to end: `curl` through a SOCKS
inbound, out through the embedded engine, to the internet.
