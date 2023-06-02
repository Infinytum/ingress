package proxy

import (
	"fmt"
	"net"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	proxyproto "github.com/pires/go-proxyproto"
)

type Proxy struct {
	// Timeout specifies an optional maximum time for the PROXY header to be received. If zero, timeout is disabled. Default is 5s.
	Timeout caddy.Duration `json:"timeout,omitempty"`

	// Allow is an optional list of CIDR ranges to allow/require PROXY headers from.
	Allow []string `json:"allow,omitempty"`

	rules []net.IPNet
}

func (Proxy) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.listeners.proxy_protocol",
		New: func() caddy.Module { return new(Proxy) },
	}
}

// UnmarshalCaddyfile sets up the listener wrapper from Caddyfile tokens. Syntax:
//
//     proxy_protocol {
//         timeout <duration>
//         allow <IPs...>
//     }
//

func (w *Proxy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		// No same-line options are supported
		if d.NextArg() {
			return d.ArgErr()
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "timeout":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("parsing proxy_protocol timeout duration: %v", err)
				}
				w.Timeout = caddy.Duration(dur)

			case "allow":
				w.Allow = append(w.Allow, d.RemainingArgs()...)

			default:
				return d.ArgErr()
			}
		}
	}
	return nil
}

func (pp *Proxy) Provision(ctx caddy.Context) error {
	pp.rules = make([]net.IPNet, 0)
	for _, s := range pp.Allow {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return fmt.Errorf("invalid subnet '%s': %w", s, err)
		}
		pp.rules = append(pp.rules, *n)
	}
	return nil
}

func (pp *Proxy) WrapListener(l net.Listener) net.Listener {
	pL := &proxyproto.Listener{
		Listener:          l,
		ReadHeaderTimeout: time.Duration(pp.Timeout),
		Policy: func(upstream net.Addr) (proxyproto.Policy, error) {
			if len(pp.Allow) == 0 {
				return proxyproto.USE, nil
			}

			var ip net.IP
			switch addr := upstream.(type) {
			case *net.UDPAddr:
				ip = addr.IP
			case *net.TCPAddr:
				ip = addr.IP
			default:
				return proxyproto.REJECT, fmt.Errorf("unsupported address type: %T", upstream.Network())
			}

			for _, rule := range pp.rules {
				if rule.Contains(ip) {
					return proxyproto.USE, nil
				}
			}

			return proxyproto.IGNORE, nil
		},
	}

	return pL
}
