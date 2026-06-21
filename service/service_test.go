package service

import (
	"testing"
)

// TestLifecycle builds a full instance (route rules + curated registries +
// embedded xray engine), starts it on an ephemeral socks port (no TUN, no
// admin), then closes it — exercising the whole New/Start/Close path and the
// outbound's engine start/stop pairing.
func TestLifecycle(t *testing.T) {
	const cfg = `{
		"log": { "level": "warn" },
		"inbounds": [
			{ "type": "socks", "tag": "in", "listen": "127.0.0.1", "listen_port": 0 }
		],
		"outbounds": [
			{ "type": "xray", "tag": "proxy", "xray": { "outbounds": [ { "protocol": "freedom", "settings": {} } ] } },
			{ "type": "direct", "tag": "direct" }
		],
		"route": {
			"final": "proxy",
			"rules": [
				{ "action": "sniff" },
				{ "protocol": "dns", "action": "hijack-dns" }
			]
		}
	}`

	instance, err := New([]byte(cfg), Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err = instance.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err = instance.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestInvalidConfigReports ensures a bad config surfaces an error instead of
// panicking.
func TestInvalidConfigReports(t *testing.T) {
	if _, err := New([]byte("{ not json"), Options{}); err == nil {
		t.Fatal("expected error for malformed config")
	}
}
