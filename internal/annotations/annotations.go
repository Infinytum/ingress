package annotations

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Annotation string

const (
	// AnnotationIngressClass is namespace for all infinytum ingress annotations
	annotationNamespace = "infinytum.ingress.kubernetes.io"

	AnnotationSSLRedirect        Annotation = annotationNamespace + "ssl-redirect"
	AnnotationBackendProtocol    Annotation = annotationNamespace + "backend-protocol"
	AnnotationInsecureSkipVerify Annotation = annotationNamespace + "insecure-skip-verify"
)

// GetAnnotation returns the value of the given annotation on the given object
func GetAnnotation(t metav1.ObjectMeta, annotation Annotation) string {
	if t.Annotations == nil {
		return ""
	}
	return t.Annotations[string(annotation)]
}

// GetAnnotation returns the value of the given annotation on the given object
func GetAnnotationOrDefault(t metav1.ObjectMeta, annotation Annotation, def string) string {
	if t.Annotations == nil || !HasAnnotation(t, annotation) {
		return def
	}
	return t.Annotations[string(annotation)]
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

// HasAnnotation checks if the given annotation exists on the given object
func HasAnnotation(t metav1.ObjectMeta, annotation Annotation) bool {
	_, ok := t.Annotations[string(annotation)]
	return ok
}
