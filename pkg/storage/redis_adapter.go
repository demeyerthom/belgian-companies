package storage

import (
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/go-redis/redis"
)

type RedisAdapter struct {
	Client redis.Client
}

func NewRedisAdapter(client redis.Client) *RedisAdapter {
	return &RedisAdapter{Client: client}
}

func (r *RedisAdapter) GetRecord(publication *model.Publication) (record *Record, err error) {

}

func (r *RedisAdapter) Close() error {
	return r.Client.Close()
}
