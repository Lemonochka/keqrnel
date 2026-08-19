// Package service is keqrnel's programmatic entry point: build a running
// instance from a config.
//
// Options.Platform is there for hosts that hand out the TUN fd and protect
// sockets themselves (Android's VpnService, via sing-box's
// adapter.PlatformInterface). Nothing in this repo supplies one — on desktop it
// stays nil and everything here is plain Go.
package service

import (
	"context"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	singservice "github.com/sagernet/sing/service"

	"github.com/Lemonochka/keqrnel/core"
)

// Options tunes how an Instance is built.
type Options struct {
	// Platform, if set, is injected so sing-box's TUN/network layer can use the
	// host platform (Android VpnService: tun fd, protected sockets, interface
	// monitor). Leave nil on desktop.
	Platform adapter.PlatformInterface
}

// Instance is a constructed, not-yet-started keqrnel core.
type Instance struct {
	box *box.Box
}

// New decodes the config through keqrnel's curated registries (so the "xray"
// outbound is understood) and constructs the Box. It does not start it.
func New(configJSON []byte, options Options) (*Instance, error) {
	ctx := core.Context(context.Background())
	if options.Platform != nil {
		ctx = singservice.ContextWith[adapter.PlatformInterface](ctx, options.Platform)
	}

	opts, err := json.UnmarshalExtendedContext[option.Options](ctx, configJSON)
	if err != nil {
		return nil, E.Cause(err, "decode config")
	}

	instance, err := box.New(box.Options{Context: ctx, Options: opts})
	if err != nil {
		return nil, E.Cause(err, "create service")
	}
	return &Instance{box: instance}, nil
}

// Start brings the instance up (opens TUN, binds inbounds, starts the embedded
// xray engine).
func (i *Instance) Start() error {
	return i.box.Start()
}

// Close tears the instance down, stopping the embedded xray engine too.
func (i *Instance) Close() error {
	return i.box.Close()
}
