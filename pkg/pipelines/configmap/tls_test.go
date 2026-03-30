package configmap

import (
	"testing"
)

func TestGenerateLEIssuer_EmptyEmail(t *testing.T) {
	ctx := &Context{AcmeEmail: ""}
	if generateLEIssuer(ctx) != nil {
		t.Error("expected nil when email is empty")
	}
}

func TestGenerateLEIssuer_WithEmail(t *testing.T) {
	ctx := &Context{AcmeEmail: "admin@example.com"}
	issuer := generateLEIssuer(ctx)
	if issuer == nil {
		t.Fatal("expected non-nil issuer")
	}
	if issuer.CA != "https://acme-v02.api.letsencrypt.org/directory" {
		t.Errorf("unexpected CA: %q", issuer.CA)
	}
	if issuer.Email != "admin@example.com" {
		t.Errorf("unexpected email: %q", issuer.Email)
	}
	if issuer.Challenges == nil {
		t.Fatal("expected challenges config")
	}
	if issuer.Challenges.HTTP.AlternatePort != 8080 {
		t.Errorf("HTTP alternate port = %d, want 8080", issuer.Challenges.HTTP.AlternatePort)
	}
	if issuer.Challenges.TLSALPN.AlternatePort != 8443 {
		t.Errorf("TLSALPN alternate port = %d, want 8443", issuer.Challenges.TLSALPN.AlternatePort)
	}
	if issuer.ExternalAccount != nil {
		t.Error("LE issuer should not have external account")
	}
}

func TestGenerateZeroSSLIssuer_EmptyEmail(t *testing.T) {
	ctx := &Context{AcmeEmail: ""}
	if generateZeroSSLIssuer(ctx) != nil {
		t.Error("expected nil when email is empty")
	}
}

func TestGenerateZeroSSLIssuer_MissingEABKeyId(t *testing.T) {
	ctx := &Context{
		AcmeEmail:     "admin@example.com",
		AcmeEABKeyId:  "",
		AcmeEABMacKey: "mac123",
	}
	if generateZeroSSLIssuer(ctx) != nil {
		t.Error("expected nil when EAB key ID is missing")
	}
}

func TestGenerateZeroSSLIssuer_MissingEABMacKey(t *testing.T) {
	ctx := &Context{
		AcmeEmail:     "admin@example.com",
		AcmeEABKeyId:  "key123",
		AcmeEABMacKey: "",
	}
	if generateZeroSSLIssuer(ctx) != nil {
		t.Error("expected nil when EAB MAC key is missing")
	}
}

func TestGenerateZeroSSLIssuer_AllSet(t *testing.T) {
	ctx := &Context{
		AcmeEmail:     "admin@example.com",
		AcmeEABKeyId:  "key123",
		AcmeEABMacKey: "mac456",
	}
	issuer := generateZeroSSLIssuer(ctx)
	if issuer == nil {
		t.Fatal("expected non-nil issuer")
	}
	if issuer.CA != "https://acme.zerossl.com/v2/DV90" {
		t.Errorf("unexpected CA: %q", issuer.CA)
	}
	if issuer.Email != "admin@example.com" {
		t.Errorf("unexpected email: %q", issuer.Email)
	}
	if issuer.ExternalAccount == nil {
		t.Fatal("expected external account")
	}
	if issuer.ExternalAccount.KeyID != "key123" {
		t.Errorf("unexpected EAB key ID: %q", issuer.ExternalAccount.KeyID)
	}
	if issuer.ExternalAccount.MACKey != "mac456" {
		t.Errorf("unexpected EAB MAC key: %q", issuer.ExternalAccount.MACKey)
	}
	if issuer.Challenges.HTTP.AlternatePort != 8080 {
		t.Errorf("HTTP alternate port = %d, want 8080", issuer.Challenges.HTTP.AlternatePort)
	}
}

func TestGenerateBuyPassIssuer_EmptyEmail(t *testing.T) {
	ctx := &Context{AcmeEmail: ""}
	if generateBuyPassIssuer(ctx) != nil {
		t.Error("expected nil when email is empty")
	}
}

func TestGenerateBuyPassIssuer_WithEmail(t *testing.T) {
	ctx := &Context{AcmeEmail: "admin@example.com"}
	issuer := generateBuyPassIssuer(ctx)
	if issuer == nil {
		t.Fatal("expected non-nil issuer")
	}
	if issuer.CA != "https://api.buypass.com/acme/directory" {
		t.Errorf("unexpected CA: %q", issuer.CA)
	}
	if issuer.Email != "admin@example.com" {
		t.Errorf("unexpected email: %q", issuer.Email)
	}
	if issuer.ExternalAccount != nil {
		t.Error("BuyPass issuer should not have external account")
	}
}
