package main

import (
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/broker/rdb"
	"github.com/engpetarmarinov/gotama/internal/config"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"github.com/engpetarmarinov/gotama/internal/worker"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.NewConfig()
	rco := rdb.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Get("REDIS_ADDR"), cfg.Get("REDIS_PORT")),
		Password: cfg.Get("REDIS_PASSWORD"),
	}

	client, ok := rco.MakeRedisClient().(redis.UniversalClient)
	if !ok {
		panic("panic casting to redis client")
	}

	clock := timeutil.NewRealClock()
	broker := rdb.NewRDB(client, clock)
	wrk := worker.NewWorker(broker, cfg, clock)
	wrk.Run()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	<-shutdown
	slog.Info("graceful shutdown...")
	err := wrk.Shutdown()
	if err != nil {
		slog.Error("error shutting down broker", "error", err)
	}

	err = broker.Close()
	if err != nil {
		slog.Error("error closing broker", "error", err)
	}
}
