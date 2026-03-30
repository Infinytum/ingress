package configmap

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeCM(data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cm", Namespace: "default"},
		Data:       data,
	}
}

func TestParseConfigMap_AllFields(t *testing.T) {
	cm := makeCM(map[string]string{
		"acmeEmail":              "admin@example.com",
		"acmeEABKeyId":           "key123",
		"acmeEABMacKey":          "mac456",
		"onDemandTLS":            "true",
		"onDemandTLSAsk":        "http://example.com/ask",
		"onDemandTLSInternalAsk": "true",
		"proxyProtocol":          "true",
		"ocspCheckInterval":      "30m",
	})

	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.AcmeEmail != "admin@example.com" {
		t.Errorf("AcmeEmail = %q, want 'admin@example.com'", ctx.AcmeEmail)
	}
	if ctx.AcmeEABKeyId != "key123" {
		t.Errorf("AcmeEABKeyId = %q, want 'key123'", ctx.AcmeEABKeyId)
	}
	if ctx.AcmeEABMacKey != "mac456" {
		t.Errorf("AcmeEABMacKey = %q, want 'mac456'", ctx.AcmeEABMacKey)
	}
	if !ctx.OnDemandTLS {
		t.Error("OnDemandTLS should be true")
	}
	if ctx.OnDemandAsk != "http://example.com/ask" {
		t.Errorf("OnDemandAsk = %q", ctx.OnDemandAsk)
	}
	if !ctx.OnDemandInternalAsk {
		t.Error("OnDemandInternalAsk should be true")
	}
	if !ctx.ProxyProtocol {
		t.Error("ProxyProtocol should be true")
	}
	if ctx.OCSPCheckInterval == 0 {
		t.Error("OCSPCheckInterval should be non-zero")
	}
}

func TestParseConfigMap_Empty(t *testing.T) {
	cm := makeCM(map[string]string{})
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.AcmeEmail != "" {
		t.Errorf("expected empty AcmeEmail, got %q", ctx.AcmeEmail)
	}
	if ctx.OnDemandTLS {
		t.Error("OnDemandTLS should default to false")
	}
	if ctx.ProxyProtocol {
		t.Error("ProxyProtocol should default to false")
	}
}

func TestParseConfigMap_BooleanFalse(t *testing.T) {
	cm := makeCM(map[string]string{
		"onDemandTLS":   "false",
		"proxyProtocol": "false",
	})
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.OnDemandTLS {
		t.Error("OnDemandTLS should be false")
	}
	if ctx.ProxyProtocol {
		t.Error("ProxyProtocol should be false")
	}
}

func TestParseConfigMap_DurationParsing(t *testing.T) {
	cm := makeCM(map[string]string{
		"ocspCheckInterval": "1h",
	})
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1h = 3600 seconds, caddy.Duration is in nanoseconds
	if ctx.OCSPCheckInterval == 0 {
		t.Error("OCSPCheckInterval should be non-zero for '1h'")
	}
}

func TestParseConfigMap_UnknownFields(t *testing.T) {
	cm := makeCM(map[string]string{
		"acmeEmail":    "test@example.com",
		"unknownField": "should be ignored",
	})
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.AcmeEmail != "test@example.com" {
		t.Errorf("AcmeEmail = %q", ctx.AcmeEmail)
	}
}

func TestParseConfigMap_NilData(t *testing.T) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.AcmeEmail != "" {
		t.Error("expected zero-value context")
	}
}

func TestParseConfigMap_ProxyProtocolAllowedIPs(t *testing.T) {
	cm := makeCM(map[string]string{
		"proxyProtocol": "true",
	})
	ctx, err := ParseConfigMap(cm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.ProxyProtocolAllowedIPs != nil && len(ctx.ProxyProtocolAllowedIPs) != 0 {
		t.Errorf("expected nil/empty ProxyProtocolAllowedIPs, got %v", ctx.ProxyProtocolAllowedIPs)
	}
}
