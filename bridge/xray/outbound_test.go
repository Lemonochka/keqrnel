package xray

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
)

// TestOutboundEngineStopsWithScope pins the engine's lifetime to the start
// scope: since sing-box 1.14.3 the Box closes scopes and never calls Close on
// an outbound, so without the registered cleanup the engine outlived the Box.
func TestOutboundEngineStopsWithScope(t *testing.T) {
	out, err := New(context.Background(), nil, log.NewNOPFactory().Logger(), "proxy", Options{
		Config: json.RawMessage(`{"outbounds":[{"protocol":"freedom","settings":{}}]}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	o := out.(*Outbound)
	scope := adapter.NewScope(context.Background(), log.NewNOPFactory().Logger())
	if err = o.Start(adapter.StartStateStart, scope); err != nil {
		t.Fatal(err)
	}
	if _, err = o.currentEngine(); err != nil {
		t.Fatalf("engine not running after start: %v", err)
	}
	if err = scope.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = o.currentEngine(); err == nil {
		t.Fatal("engine still running after the scope closed")
	}
}
