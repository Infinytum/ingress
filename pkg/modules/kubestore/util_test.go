package kubestore

import (
	"testing"
)

func TestCleanKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"slashes replaced", "certificates/acme/example.com", "certificates.acme.example.com"},
		{"dots replaced", "foo.bar.baz", "foo.bar.baz"},
		{"mixed special chars", "a/b.c@d!e", "a.b.c.d.e"},
		{"hyphens preserved", "my-secret-name", "my-secret-name"},
		{"alphanumeric preserved", "abc123", "abc123"},
		{"empty string", "", ""},
		{"consecutive specials collapsed", "a//b", "a.b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanKey(tt.input)
			if got != tt.want {
				t.Errorf("cleanKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateSecretName_SecretPrefix(t *testing.T) {
	// secret/namespace/name/datakey -> returns the name part
	got := generateSecretName("secret/my-namespace/my-tls-secret/tls.crt")
	if got != "my-tls-secret" {
		t.Errorf("expected 'my-tls-secret', got %q", got)
	}
}

func TestGenerateSecretName_NormalKey(t *testing.T) {
	got := generateSecretName("certificates/acme/example.com")
	want := "tls-certificates.acme.example.com--acme"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestGenerateSecretName_SimpleKey(t *testing.T) {
	got := generateSecretName("mykey")
	want := "tls-mykey--acme"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestGetNamespace_SecretPrefix(t *testing.T) {
	got := getNamespace("secret/prod/cert/tls.crt", "default")
	if got != "prod" {
		t.Errorf("expected 'prod', got %q", got)
	}
}

func TestGetNamespace_NoSecretPrefix(t *testing.T) {
	got := getNamespace("certificates/acme/foo", "default")
	if got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestGetDataKey_SecretPrefix(t *testing.T) {
	got := getDataKey("secret/ns/name/tls.crt")
	if got != "tls.crt" {
		t.Errorf("expected 'tls.crt', got %q", got)
	}
}

func TestGetDataKey_SecretPrefixDifferentKey(t *testing.T) {
	got := getDataKey("secret/ns/name/tls.key")
	if got != "tls.key" {
		t.Errorf("expected 'tls.key', got %q", got)
	}
}

func TestGetDataKey_NormalKey(t *testing.T) {
	got := getDataKey("certificates/acme/foo")
	if got != dataKey {
		t.Errorf("expected %q, got %q", dataKey, got)
	}
}
