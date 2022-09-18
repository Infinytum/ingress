package ingress

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/infinytum/reactive"
	networkingv1 "k8s.io/api/networking/v1"
)

type ContextMode int

const (
	// ContextModeConfigure is the context mode for adding or updating an ingress
	ContextModeConfigure ContextMode = iota
	// ContextModeDelete is the context mode for removing an ingress
	ContextModeDelete
)

type Context struct {
	Ingress *networkingv1.Ingress
	Mode    ContextMode
}

func (s Context) IsNewer(routeIdentifier string) bool {
	parts := strings.Split(routeIdentifier, "--")
	resourceVersion := parts[2]
	generation, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}

	if s.Ingress.Generation > int64(generation) {
		return true
	}

	// Only apply resource version updates when generation is the same
	// This is to prevent race conditions.
	if s.Ingress.Generation == int64(generation) && s.Ingress.ResourceVersion != resourceVersion {
		return true
	}
	return false
}

func (s Context) RouteIdentifier() string {
	return fmt.Sprintf("%s--%d--%s", s.Ingress.UID, s.Ingress.Generation, s.Ingress.ResourceVersion)
}

type GlobalContext struct {
	Context
	Routes []*caddyhttp.Route
}

type SpecificContext struct {
	Context
	Path  networkingv1.HTTPIngressPath
	Route *caddyhttp.Route
	Rule  networkingv1.IngressRule
}

// GlobalPipe is a pipe which takes the output of a previous GlobalPipe, works
// with received input and then again produces an output for the next GlobalPipe
func GlobalPipe(f func(ctx *GlobalContext, errs []error) []error) reactive.Pipe {
	return reactive.Pipe(func(parent reactive.Observable, next reactive.Subjectable) {
		parent.Subscribe((func(ctx *GlobalContext, errs []error) {
			errs = f(ctx, errs)
			next.Next(ctx, errs)
		}))
	})
}

// SpecificPipe is a pipe which takes the output of a previous SpecificPipe, works
// with received input and then again produces an output for the next SpecificPipe
func SpecificPipe(f func(ctx *SpecificContext, errs []error) []error) reactive.Pipe {
	return reactive.Pipe(func(parent reactive.Observable, next reactive.Subjectable) {
		parent.Subscribe((func(ctx *SpecificContext, errs []error) {
			errs = f(ctx, errs)
			next.Next(ctx, errs)
		}))
	})
}
