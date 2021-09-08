//go:build linux || darwin
// +build linux darwin

package proc

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tal-tech/go-zero/core/logx"
)

//const timeFormat = "0102150405"

func init() {
	go func() {
		var profiler Stopper

		// https://golang.org/pkg/os/signal/#Notify
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP, syscall.SIGTERM)

		for {
			v := <-signals
			switch v {
			case syscall.SIGUSR1:
				logx.SetLevel(logx.DebugLevel)
				logx.Alert("Signal SIGUSR1 received, set log level to DEBUG")
			case syscall.SIGUSR2:
				logx.SetLevel(logx.InfoLevel)
				logx.Alert("Signal SIGUSR2 received, set log level to INFO")
			case syscall.SIGHUP:
				dumpGoroutines()
				//if profiler == nil {
				//	profiler = StartProfile()
				//} else {
				//	profiler.Stop()
				//	profiler = nil
				//}
			case syscall.SIGTERM:
				select {
				case <-done:
					// already closed
				default:
					close(done)
				}
				gracefulStop(signals)
			default:
				logx.Error("Got unregistered signal:", v)
			}
		}
	}()
}
