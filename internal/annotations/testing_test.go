package annotations

import (
	"os"
	"testing"

	"github.com/infinytum/ingress/internal/service"
	"github.com/infinytum/injector"
)

func TestMain(m *testing.M) {
	// Pre-register IngressConfig to prevent flag.Parse() from being
	// called by the injector during init(), which conflicts with test flags.
	injector.Singleton(func() service.IngressConfig {
		return service.IngressConfig{
			ClassName:        "test",
			ConfigMap:        "test-config",
			EnableHSTS:       true,
			NginxAnnotations: false,
		}
	})

	os.Exit(m.Run())
}
