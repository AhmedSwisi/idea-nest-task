package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var ctx context.Context
var client *mongo.Client
var rdb *redis.Client

func Init(uri string, databaseName string) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	var err error

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	db = client.Database(databaseName)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
}

func InitRedis() {
	var context = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	err := rdb.Set(context, "key", "value", 0).Err()

	if err != nil {
		panic(err)
	}
}

func GetRedis() *redis.Client {
	return rdb
}

func Close() error {
	return client.Disconnect(context.Background())
}

func GetDB() *mongo.Database {
	return db
}

func RedisError() error {
	return redis.Nil
}
