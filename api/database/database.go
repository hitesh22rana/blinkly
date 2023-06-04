package database

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var Addr = os.Getenv("DB_ADDR")
var Password = os.Getenv("DB_PASS")

func CreateClient(dbNo int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       dbNo,
	})

	return rdb
}
