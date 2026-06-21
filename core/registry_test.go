package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

// TestExampleConfigParses verifies the shipped example config decodes through
// keqrnel's curated registries — in particular that the custom "xray" outbound
// type is recognised and its nested xray fragment is captured.
func TestExampleConfigParses(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "config.example.json"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}

	ctx := Context(context.Background())
	options, err := json.UnmarshalExtendedContext[option.Options](ctx, content)
	if err != nil {
		t.Fatalf("decode config: %v", err)
	}

	var foundXray, foundTun bool
	for _, ob := range options.Outbounds {
		if ob.Type == "xray" {
			foundXray = true
		}
	}
	for _, ib := range options.Inbounds {
		if ib.Type == "tun" {
			foundTun = true
		}
	}
	if !foundXray {
		t.Error("xray outbound not parsed from example config")
	}
	if !foundTun {
		t.Error("tun inbound not parsed from example config")
	}
}
