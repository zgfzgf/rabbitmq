package mqengine

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisSnapshotStore struct {
	productId   string
	redisClient *redis.Client
}

func NewRedisSnapshotStore(productId string) SnapshotStore {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       0,
	})
	return &RedisSnapshotStore{
		productId:   productId,
		redisClient: redisClient,
	}
}

func (s *RedisSnapshotStore) Store(snapshot []byte) error {
	return s.redisClient.Set(SnapshotPrefix+s.productId, snapshot, SnapshotHour*time.Hour).Err()
}

func (s *RedisSnapshotStore) GetLatest() ([]byte, error) {
	ret, err := s.redisClient.Get(SnapshotPrefix + s.productId).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return ret, err
}
