package ingress

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func makeContext(uid string, generation int64, resourceVersion string) Context {
	return Context{
		Ingress: &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				UID:             types.UID(uid),
				Generation:      generation,
				ResourceVersion: resourceVersion,
			},
		},
	}
}

func TestRouteIdentifier(t *testing.T) {
	ctx := makeContext("abc-123", 5, "99999")
	got := ctx.RouteIdentifier()
	want := "abc-123--5--99999"
	if got != want {
		t.Errorf("RouteIdentifier() = %q, want %q", got, want)
	}
}

func TestIsNewer_HigherGeneration(t *testing.T) {
	ctx := makeContext("uid", 3, "100")
	if !ctx.IsNewer("uid--2--100") {
		t.Error("expected newer when generation is higher")
	}
}

func TestIsNewer_SameGenerationDifferentResourceVersion(t *testing.T) {
	ctx := makeContext("uid", 2, "200")
	if !ctx.IsNewer("uid--2--100") {
		t.Error("expected newer when same generation but different resourceVersion")
	}
}

func TestIsNewer_ExactMatch(t *testing.T) {
	ctx := makeContext("uid", 2, "100")
	if ctx.IsNewer("uid--2--100") {
		t.Error("expected not newer when generation and resourceVersion match")
	}
}

func TestIsNewer_OlderGeneration(t *testing.T) {
	ctx := makeContext("uid", 1, "100")
	if ctx.IsNewer("uid--2--100") {
		t.Error("expected not newer when generation is lower")
	}
}

func TestIsNewer_MalformedGeneration(t *testing.T) {
	ctx := makeContext("uid", 1, "100")
	// Non-numeric generation in routeIdentifier should return true (Atoi error path)
	if !ctx.IsNewer("uid--notanumber--100") {
		t.Error("expected newer when routeIdentifier has non-numeric generation")
	}
}

func TestIsNewer_GenerationZero(t *testing.T) {
	ctx := makeContext("uid", 0, "50")
	if ctx.IsNewer("uid--0--50") {
		t.Error("expected not newer when both are generation 0 with same resourceVersion")
	}
}

func TestIsNewer_SameGenerationSameResourceVersion(t *testing.T) {
	ctx := makeContext("uid", 5, "12345")
	if ctx.IsNewer("uid--5--12345") {
		t.Error("expected not newer when generation and resourceVersion both match")
	}
}
