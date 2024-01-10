package db

import (
	"context"
	"encoding/json"

	"github.com/kafkago/db/models"
	"github.com/redis/go-redis/v9"
)

type Redis[M models.Keyer] struct {
	rdb *redis.Client
}

func NewRedis[M models.Keyer] (rdb *redis.Client) Redis[M] {
	r := Redis[M]{rdb: rdb}
	return r
}

func (r Redis[M]) Save(ctx context.Context, k M) error {
	b, _ := json.Marshal(k)
	return r.rdb.Set(ctx, k.Key(), b, 0).Err()
}

func (r Redis[M]) Get(ctx context.Context, key string) (M, error) {
	var m M
	b, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return m, err
	}
	json.Unmarshal(b, &m)
	return m, nil

}