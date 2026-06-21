// Package localdns provides keqrnel's own "local" DNS transport. sing-box's
// bundled dns/transport/local pulls platform glue (and is currently broken on
// Windows: resolv_windows.go calls a renamed sing-tun method). libbox replaces
// it with a platform transport for the same reason; we replace it with a small
// cross-platform one backed by Go's system resolver. sing-box always needs a
// "local" transport registered: it's the default DNS fallback (box.go), used
// e.g. to resolve a proxy server's own domain.
package localdns

import (
	"context"
	"net"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"

	mDNS "github.com/miekg/dns"
)

// RegisterTransport registers the "local" DNS transport type.
func RegisterTransport(registry *dns.TransportRegistry) {
	dns.RegisterTransport[option.LocalDNSServerOptions](registry, C.DNSTypeLocal, NewTransport)
}

var _ adapter.DNSTransport = (*Transport)(nil)

// Transport answers A/AAAA queries via the host's system resolver.
type Transport struct {
	dns.TransportAdapter
	resolver *net.Resolver
}

// NewTransport constructs the local DNS transport.
func NewTransport(ctx context.Context, logger log.ContextLogger, tag string, options option.LocalDNSServerOptions) (adapter.DNSTransport, error) {
	return &Transport{
		TransportAdapter: dns.NewTransportAdapterWithLocalOptions(C.DNSTypeLocal, tag, options),
		resolver:         &net.Resolver{},
	}, nil
}

func (t *Transport) Start(stage adapter.StartStage) error { return nil }

func (t *Transport) Close() error { return nil }

func (t *Transport) Reset() {}

// Exchange resolves A/AAAA through the system resolver; other query types get
// an empty NOERROR response (this transport is a last-resort fallback).
func (t *Transport) Exchange(ctx context.Context, message *mDNS.Msg) (*mDNS.Msg, error) {
	if len(message.Question) == 0 {
		return dns.FixedResponseStatus(message, mDNS.RcodeFormatError), nil
	}
	question := message.Question[0]

	var network string
	switch question.Qtype {
	case mDNS.TypeA:
		network = "ip4"
	case mDNS.TypeAAAA:
		network = "ip6"
	default:
		return dns.FixedResponse(message.Id, question, nil, C.DefaultDNSTTL), nil
	}

	addresses, err := t.resolver.LookupNetIP(ctx, network, dns.FqdnToDomain(question.Name))
	if err != nil {
		return dns.FixedResponseStatus(message, mDNS.RcodeServerFailure), nil
	}
	return dns.FixedResponse(message.Id, question, addresses, C.DefaultDNSTTL), nil
}
