package database

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {
	opt, err := redis.ParseURL(os.Getenv("DB_ADDR"))
	if err != nil {
		log.Fatal("Redis URL parse error:", err)
	}
	opt.DB = dbNo

	rdb := redis.NewClient(opt)
	return rdb
}
