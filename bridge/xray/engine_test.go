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

// TestEngineUDPRoundTrip proves the connected-UDP path (DialContext("udp")) —
// the one DNS-over-proxy uses — works through the embedded engine.
func TestEngineUDPRoundTrip(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			pc.WriteTo(buf[:n], addr)
		}
	}()

	const cfg = `{ "outbounds": [{ "protocol": "freedom", "settings": {} }] }`
	engine, err := NewEngine([]byte(cfg))
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	defer engine.Close()

	addr := pc.LocalAddr().(*net.UDPAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := engine.DialUDP(ctx, "127.0.0.1", uint16(addr.Port))
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer conn.Close()

	want := []byte("udp via keqrnel")
	if _, err = conn.Write(want); err != nil {
		t.Fatalf("write: %v", err)
	}
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	got := make([]byte, len(want))
	n, err := conn.Read(got)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got[:n]) != string(want) {
		t.Fatalf("udp round-trip mismatch: got %q want %q", got[:n], want)
	}
}
