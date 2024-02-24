package rdb

import (
	"context"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"github.com/redis/go-redis/v9"
)

type RDB struct {
	broker.Broker
	client redis.UniversalClient
	clock  timeutil.Clock
}

func NewRDB(client redis.UniversalClient) *RDB {
	return &RDB{
		client: client,
		clock:  timeutil.NewRealClock(),
	}
}

func (r *RDB) Ping() error {
	return r.client.Ping(context.Background()).Err()
}

func (r *RDB) Close() error {
	return r.client.Close()
}
