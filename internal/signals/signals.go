package signals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/infinytum/injector"
)

type Signal chan struct{}

const (
	STOP      = "STOP"
	INTERRUPT = "INTERRUPT"
)

func init() {
	injector.Singleton(func() Signal {
		c := make(Signal, 1)
		go func() {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
			<-sigs
			c <- struct{}{}
		}()
		return c
	}, INTERRUPT)
	injector.Singleton(func() Signal {
		return make(Signal, 1)
	}, STOP)
}
