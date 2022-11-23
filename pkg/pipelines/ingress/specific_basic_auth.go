package ingress

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/infinytum/ingress/internal/annotations"
	"github.com/infinytum/reactive"
)

type Authentication struct {
	Providers AuthenticationProviders `json:"providers,omitempty"`
}

type AuthenticationProviders struct {
	HttpBasic AuthenticationHttpBasic `json:"http_basic,omitempty"`
}

type AuthenticationHttpBasic struct {
	Accounts []AuthenticationHttpBasicAccount `json:"accounts,omitempty"`
}

type AuthenticationHttpBasicAccount struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Salt     string `json:"salt,omitempty"`
}

// SpecificBasicAuth configures a route to authenticate requests using basic auth
func SpecificBasicAuth() reactive.Pipe {
	return SpecificPipe(func(ctx *SpecificContext, errs []error) []error {
		// If there are already errors or the ingress is being deleted, skip this pipe
		if len(errs) > 0 || ctx.Mode == ContextModeDelete {
			return errs
		}

		// Configure rewrite target
		username := annotations.GetAnnotation(ctx.Ingress.ObjectMeta, annotations.AnnotationBasicUsername)
		password := annotations.GetAnnotation(ctx.Ingress.ObjectMeta, annotations.AnnotationBasicPassword)
		if username == "" || password == "" {
			return errs
		}

		handlerModule := caddyconfig.JSONModuleObject(
			Authentication{
				Providers: AuthenticationProviders{
					HttpBasic: AuthenticationHttpBasic{
						Accounts: []AuthenticationHttpBasicAccount{
							{
								Username: username,
								Password: password,
								Salt:     annotations.GetAnnotation(ctx.Ingress.ObjectMeta, annotations.AnnotationBasicSalt),
							},
						},
					},
				},
			},
			"handler",
			"authentication",
			nil,
		)

		ctx.Route.HandlersRaw = append(ctx.Route.HandlersRaw, handlerModule)
		return errs
	})
}
