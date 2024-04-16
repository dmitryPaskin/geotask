package cache

import (
	"github.com/go-redis/redis"
)

func NewRedisClient(host, port string) *redis.Client {
	// реализуйте создание клиента для Redis
	client := redis.NewClient(&redis.Options{
		Network: host,
		Addr:    port,
	})
	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}
	return client
}
