// Command keqrnel runs the unified core: sing-box host with an embedded xray
// engine. Usage: keqrnel [config.json]  (defaults to ./config.json).
package main

import (
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sagernet/sing-box/log"

	"github.com/Lemonochka/keqrnel/service"
)

func main() {
	configPath := parseConfigPath(os.Args[1:])

	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("read config: ", err)
	}

	instance, err := service.New(content, service.Options{})
	if err != nil {
		log.Fatal(err)
	}

	if err = instance.Start(); err != nil {
		log.Fatal("start: ", err)
	}
	log.Info("keqrnel started")

	stop := make(chan struct{}, 1)
	requestStop := func() {
		select {
		case stop <- struct{}{}:
		default:
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		requestStop()
	}()

	// Graceful shutdown when the parent closes our stdin. On Windows there is no
	// real SIGTERM — Dart's Process.kill is a hard TerminateProcess that gives
	// sing-box no chance to revert the TUN adapter, routes and DNS, which can
	// leave the network/TUN broken after a disconnect or crash. The desktop
	// launcher wires up a stdin pipe and closes it to ask for a clean stop.
	//
	// Only watch stdin when it is an actual pipe. Under Android's fork+exec stdin
	// is /dev/null or closed, so reading it returns EOF immediately — that must
	// NOT be taken as a shutdown request (it would kill the tunnel the instant it
	// starts).
	go func() {
		info, err := os.Stdin.Stat()
		if err != nil || info.Mode()&os.ModeNamedPipe == 0 {
			return
		}
		io.Copy(io.Discard, os.Stdin)
		requestStop()
	}()

	<-stop
	log.Info("keqrnel shutting down")
	if err = instance.Close(); err != nil {
		log.Error("close: ", err)
	}
}

// parseConfigPath resolves the config path from CLI args, accepting both
// keqrnel's own `keqrnel <config>` form and xray's `run -c <config>` form so the
// binary is a drop-in replacement for xray in keqdroid's launcher (which
// fork+execs the core the same way regardless of engine).
func parseConfigPath(args []string) string {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "-config", "--config":
			if i+1 < len(args) {
				return args[i+1]
			}
		case "run":
			continue // xray subcommand, ignore
		default:
			if !strings.HasPrefix(args[i], "-") {
				return args[i]
			}
		}
	}
	return "config.json"
}
