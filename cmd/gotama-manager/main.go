package main

import (
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/manager"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.NewConfig()
	manager.NewManager().Run(cfg)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown
	//TODO: graceful shutdown
}
