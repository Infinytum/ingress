package configmap

import (
	"reflect"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/infinytum/reactive"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type Context struct {
	// ACME configuration
	AcmeCA        string `json:"acmeCA,omitempty"`
	AcmeEABKeyId  string `json:"acmeEABKeyId,omitempty"`
	AcmeEABMacKey string `json:"acmeEABMacKey,omitempty"`
	AcmeEmail     string `json:"acmeEmail,omitempty"`

	// OnDemand TLS configuration
	OnDemandTLS               bool           `json:"onDemandTLS,omitempty"`
	OnDemandRateLimitInterval caddy.Duration `json:"onDemandTLSRateLimitInterval,omitempty"`
	OnDemandRateLimitBurst    int            `json:"onDemandTLSRateLimitBurst,omitempty"`
	OnDemandAsk               string         `json:"onDemandTLSAsk,omitempty"`
	OnDemandInternalAsk       bool           `json:"onDemandTLSInternalAsk,omitempty"`

	// Enable PROXY protocol support
	ProxyProtocol           bool     `json:"proxyProtocol,omitempty"`
	ProxyProtocolAllowedIPs []string `json:"proxyProtocolAllowedIPs,omitempty"`

	// General TLS Configuration
	OCSPCheckInterval caddy.Duration `json:"ocspCheckInterval,omitempty"`
}

func stringToCaddyDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(caddy.Duration(time.Second)) {
			return data, nil
		}
		return caddy.ParseDuration(data.(string))
	}
}

func ParseConfigMap(cm *v1.ConfigMap) (*Context, error) {
	// parse configmap
	cfgMap := Context{}
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           &cfgMap,
		TagName:          "json",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToCaddyDurationHookFunc(),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error creating decoder")
	}
	err = decoder.Decode(cm.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error parsing configmap")
	}

	return &cfgMap, nil
}

// Pipe is a pipe which takes the output of a previous Pipe, works
// with received input and then again produces an output for the next Pipe
func Pipe(f func(ctx *Context, errs []error) []error) reactive.Pipe {
	return reactive.Pipe(func(parent reactive.Observable, next reactive.Subjectable) {
		parent.Subscribe((func(ctx *Context, errs []error) {
			errs = f(ctx, errs)
			next.Next(ctx, errs)
		}))
	})
}
