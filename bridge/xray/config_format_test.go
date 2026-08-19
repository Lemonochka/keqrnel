package xray

import "testing"

// Dummy but syntactically valid credentials: xray validates the shape of these at
// config-build time, so placeholders like "YOUR_UUID" would not get us to the point
// the test is about.
const (
	testUUID      = "b831381d-6324-4d53-ad4f-8cda48b30811"
	testPublicKey = "jNXHt1yRo0vDuchQlIP6Z0ZvjT3KtzVI-T4E7RoLJS0"
	testShortID   = "0123456789abcdef"
)

// TestVLESSFlatSettings covers the outbound form current xray writes: server and
// credentials directly under "settings", no "vnext" wrapper.
func TestVLESSFlatSettings(t *testing.T) {
	cfg := `{
		"outbounds": [{
			"protocol": "vless",
			"settings": {
				"address": "example.com",
				"port": 443,
				"id": "` + testUUID + `",
				"encryption": "none",
				"flow": "xtls-rprx-vision"
			},
			"streamSettings": {
				"network": "tcp",
				"security": "reality",
				"realitySettings": {
					"serverName": "www.example.com",
					"fingerprint": "chrome",
					"publicKey": "` + testPublicKey + `",
					"shortId": "` + testShortID + `"
				}
			}
		}]
	}`

	engine, err := NewEngine([]byte(cfg))
	if err != nil {
		t.Fatalf("flat settings rejected: %v", err)
	}
	engine.Close()
}

// TestVLESSVnextSettings covers the older "vnext" form, which xray still accepts.
// Configs in the wild are full of it, and the fragment is passed through untouched,
// so it has to keep working.
func TestVLESSVnextSettings(t *testing.T) {
	cfg := `{
		"outbounds": [{
			"protocol": "vless",
			"settings": {
				"vnext": [{
					"address": "example.com",
					"port": 443,
					"users": [{ "id": "` + testUUID + `", "encryption": "none", "flow": "xtls-rprx-vision" }]
				}]
			},
			"streamSettings": { "network": "tcp", "security": "none" }
		}]
	}`

	engine, err := NewEngine([]byte(cfg))
	if err != nil {
		t.Fatalf("vnext settings rejected: %v", err)
	}
	engine.Close()
}
