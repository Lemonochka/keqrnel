// Command keqrnel runs the unified core: sing-box host with an embedded xray
// engine. Usage: keqrnel [config.json]  (defaults to ./config.json).
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sagernet/sing-box/log"

	"github.com/keqdroid/keqrnel/service"
)

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

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
