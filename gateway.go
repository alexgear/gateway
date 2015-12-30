package main

import (
	"log"

	"github.com/alexgear/gateway/api"
	"github.com/alexgear/gateway/config"
	"github.com/alexgear/gateway/gservices"
	"github.com/alexgear/gateway/worker"
)

var err error

func main() {
	err := config.Init("config.toml")
	if err != nil {
		log.Fatalf(err.Error())
	}
	cfg := config.GetConfig()

	err = gservices.Init()
	if err != nil {
		log.Fatalf(err.Error())
	}

	worker.InitWorker()
	err = api.InitServer(cfg.ServerHost, cfg.ServerPort)
	if err != nil {
		log.Fatalf("main: Error starting server: %s", err.Error())
	}
}
