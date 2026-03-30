package annotations

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reset namespaces to a known state for testing (avoid injector dependency)
func init() {
	namespaces = []string{annotationNamespace}
}

func meta(annots map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        "test-ingress",
		Namespace:   "default",
		Annotations: annots,
	}
}

func fqn(annotation Annotation) string {
	return annotationNamespace + "/" + string(annotation)
}

// --- GetAnnotation ---

func TestGetAnnotation_Found(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationBackendProtocol): "https"})
	got := GetAnnotation(m, AnnotationBackendProtocol)
	if got != "https" {
		t.Errorf("expected 'https', got %q", got)
	}
}

func TestGetAnnotation_Missing(t *testing.T) {
	m := meta(map[string]string{})
	got := GetAnnotation(m, AnnotationBackendProtocol)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestGetAnnotation_NilAnnotations(t *testing.T) {
	m := metav1.ObjectMeta{Name: "test"}
	got := GetAnnotation(m, AnnotationBackendProtocol)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestGetAnnotation_NginxNamespace(t *testing.T) {
	// Add nginx namespace temporarily
	namespaces = append(namespaces, nginxNamespace)
	defer func() { namespaces = []string{annotationNamespace} }()

	m := meta(map[string]string{
		nginxNamespace + "/backend-protocol": "grpc",
	})
	got := GetAnnotation(m, AnnotationBackendProtocol)
	if got != "grpc" {
		t.Errorf("expected 'grpc', got %q", got)
	}
}

func TestGetAnnotation_PrimaryNamespaceTakesPrecedence(t *testing.T) {
	namespaces = append(namespaces, nginxNamespace)
	defer func() { namespaces = []string{annotationNamespace} }()

	m := meta(map[string]string{
		fqn(AnnotationBackendProtocol):                  "https",
		nginxNamespace + "/" + string(AnnotationBackendProtocol): "http",
	})
	got := GetAnnotation(m, AnnotationBackendProtocol)
	if got != "https" {
		t.Errorf("expected primary namespace value 'https', got %q", got)
	}
}

// --- GetAnnotationOrDefault ---

func TestGetAnnotationOrDefault_Found(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationBackendProtocol): "https"})
	got := GetAnnotationOrDefault(m, AnnotationBackendProtocol, "http")
	if got != "https" {
		t.Errorf("expected 'https', got %q", got)
	}
}

func TestGetAnnotationOrDefault_Missing(t *testing.T) {
	m := meta(map[string]string{})
	got := GetAnnotationOrDefault(m, AnnotationBackendProtocol, "http")
	if got != "http" {
		t.Errorf("expected default 'http', got %q", got)
	}
}

// --- GetAnnotationBool ---

func TestGetAnnotationBool_True(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationSSLRedirect): "true"})
	if !GetAnnotationBool(m, AnnotationSSLRedirect, false) {
		t.Error("expected true")
	}
}

func TestGetAnnotationBool_False(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationSSLRedirect): "false"})
	if GetAnnotationBool(m, AnnotationSSLRedirect, true) {
		t.Error("expected false")
	}
}

func TestGetAnnotationBool_Missing_DefaultTrue(t *testing.T) {
	m := meta(map[string]string{})
	if !GetAnnotationBool(m, AnnotationSSLRedirect, true) {
		t.Error("expected default true")
	}
}

func TestGetAnnotationBool_Missing_DefaultFalse(t *testing.T) {
	m := meta(map[string]string{})
	if GetAnnotationBool(m, AnnotationSSLRedirect, false) {
		t.Error("expected default false")
	}
}

func TestGetAnnotationBool_NonTrueString(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationSSLRedirect): "yes"})
	if GetAnnotationBool(m, AnnotationSSLRedirect, true) {
		t.Error("expected false for non-'true' string")
	}
}

// --- GetAnnotationInt ---

func TestGetAnnotationInt_Valid(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationKeepAlive): "512"})
	got := GetAnnotationInt(m, AnnotationKeepAlive, 1024)
	if got != 512 {
		t.Errorf("expected 512, got %d", got)
	}
}

func TestGetAnnotationInt_Invalid(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationKeepAlive): "abc"})
	got := GetAnnotationInt(m, AnnotationKeepAlive, 1024)
	if got != 1024 {
		t.Errorf("expected default 1024, got %d", got)
	}
}

func TestGetAnnotationInt_Missing(t *testing.T) {
	m := meta(map[string]string{})
	got := GetAnnotationInt(m, AnnotationKeepAlive, 1024)
	if got != 1024 {
		t.Errorf("expected default 1024, got %d", got)
	}
}

func TestGetAnnotationInt_Zero(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationKeepAlive): "0"})
	got := GetAnnotationInt(m, AnnotationKeepAlive, 1024)
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// --- GetAnnotationList ---

func TestGetAnnotationList_JSONArray(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationProxyHTTPVersion): `["1.1","2"]`})
	got := GetAnnotationList(m, AnnotationProxyHTTPVersion, nil)
	if len(got) != 2 || got[0] != "1.1" || got[1] != "2" {
		t.Errorf("expected [1.1, 2], got %v", got)
	}
}

func TestGetAnnotationList_CSV(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationTrustedProxies): "10.0.0.0/8,172.16.0.0/12"})
	got := GetAnnotationList(m, AnnotationTrustedProxies, nil)
	if len(got) != 2 || got[0] != "10.0.0.0/8" || got[1] != "172.16.0.0/12" {
		t.Errorf("expected 2 CIDRs, got %v", got)
	}
}

func TestGetAnnotationList_SingleValue(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationTrustedProxies): "10.0.0.0/8"})
	got := GetAnnotationList(m, AnnotationTrustedProxies, nil)
	if len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Errorf("expected [10.0.0.0/8], got %v", got)
	}
}

func TestGetAnnotationList_Missing(t *testing.T) {
	m := meta(map[string]string{})
	def := []string{"default"}
	got := GetAnnotationList(m, AnnotationTrustedProxies, def)
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("expected default, got %v", got)
	}
}

func TestGetAnnotationList_MalformedJSON(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationProxyHTTPVersion): `[invalid`})
	def := []string{"1.1"}
	got := GetAnnotationList(m, AnnotationProxyHTTPVersion, def)
	if len(got) != 1 || got[0] != "1.1" {
		t.Errorf("expected default on malformed JSON, got %v", got)
	}
}

// --- HasAnnotation ---

func TestHasAnnotation_Present(t *testing.T) {
	m := meta(map[string]string{fqn(AnnotationSSLRedirect): "true"})
	if !HasAnnotation(m, AnnotationSSLRedirect) {
		t.Error("expected true")
	}
}

func TestHasAnnotation_Missing(t *testing.T) {
	m := meta(map[string]string{})
	if HasAnnotation(m, AnnotationSSLRedirect) {
		t.Error("expected false")
	}
}
