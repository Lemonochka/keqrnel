package xray

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

// TestEngineRoundTrip proves the embedded xray engine actually dispatches a
// connection end-to-end: a local TCP echo server, reached through core.Dial via
// a minimal xray "freedom" outbound (which dials the destination directly).
func TestEngineRoundTrip(t *testing.T) {
	// Local echo server.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.Copy(conn, conn)
	}()

	// Minimal xray config: a single freedom outbound is enough to exercise the
	// dispatcher -> routing -> outbound path that VLESS/Hysteria2 also use.
	const cfg = `{
		"outbounds": [{ "protocol": "freedom", "tag": "direct", "settings": {} }]
	}`

	engine, err := NewEngine([]byte(cfg))
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	defer engine.Close()

	addr := ln.Addr().(*net.TCPAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := engine.DialTCP(ctx, "127.0.0.1", uint16(addr.Port))
	if err != nil {
		t.Fatalf("DialTCP: %v", err)
	}
	defer conn.Close()

	want := []byte("keqrnel works")
	if _, err = conn.Write(want); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := make([]byte, len(want))
	if _, err = io.ReadFull(conn, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, want)
	}
}
