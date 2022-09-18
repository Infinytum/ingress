package annotations

import (
	"encoding/json"
	"strings"

	"github.com/go-mojito/mojito/log"
	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Annotation string

const (
	annotationNamespace = "infinytum.ingress.kubernetes.io"
	nginxNamespace      = "nginx.ingress.kubernetes.io"

	AnnotationSSLRedirect        Annotation = "ssl-redirect"
	AnnotationBackendProtocol    Annotation = "backend-protocol"
	AnnotationInsecureSkipVerify Annotation = "insecure-skip-verify"
	AnnotationProxyHTTPVersion   Annotation = "proxy-http-version"
)

var namespaces = []string{annotationNamespace}

func init() {
	injector.MustCall(func(config service.IngressConfig) {
		if config.NginxAnnotations {
			namespaces = append(namespaces, nginxNamespace)
		}
	})
}

// GetAnnotation returns the value of the given annotation on the given object
func GetAnnotation(t metav1.ObjectMeta, annotation Annotation) string {
	if t.Annotations == nil {
		return ""
	}
	for _, ns := range namespaces {
		if val, ok := t.Annotations[ns+"/"+string(annotation)]; ok {
			return val
		}
	}
	return ""
}

// GetAnnotation returns the value of the given annotation on the given object
func GetAnnotationOrDefault(t metav1.ObjectMeta, annotation Annotation, def string) string {
	if t.Annotations == nil || !HasAnnotation(t, annotation) {
		return def
	}
	return GetAnnotation(t, annotation)
}

// GetAnnotationBool returns whether the annotation has the string value "true"
// or if the annotation does not exist, the specified default
func GetAnnotationBool(t metav1.ObjectMeta, annotation Annotation, def bool) bool {
	val := GetAnnotation(t, annotation)
	if val == "" {
		return def
	}
	return val == "true"
}

// GetAnnotationList returns the list of values specified in the annotation
// or if the annotation does not exist, the specified default
func GetAnnotationList(t metav1.ObjectMeta, annotation Annotation, def []string) (res []string) {
	val := GetAnnotation(t, annotation)
	if val == "" {
		return def
	}
	if strings.HasPrefix(val, "[") {
		if err := json.Unmarshal([]byte(val), &res); err != nil {
			log.
				Field("ingress", t.Name).
				Field("namespace", t.Namespace).
				Field("annotation", annotation).Errorf("Error parsing annotation: %s", err)
			return def
		}
		return res
	}

	if strings.Contains(val, ",") {
		return strings.Split(val, ",")
	}

	return []string{val}
}

// HasAnnotation checks if the given annotation exists on the given object
func HasAnnotation(t metav1.ObjectMeta, annotation Annotation) bool {
	for _, ns := range namespaces {
		if _, ok := t.Annotations[ns+"/"+string(annotation)]; ok {
			return true
		}
	}
	return false
}
