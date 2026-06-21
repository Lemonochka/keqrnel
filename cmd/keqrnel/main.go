// Command keqrnel runs the unified core: sing-box host with an embedded xray
// engine. Usage: keqrnel [config.json]  (defaults to ./config.json).
package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sagernet/sing-box/log"

	"github.com/keqdroid/keqrnel/service"
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
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
