package database

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	RedisClient *redis.Client
	MongoClient *mongo.Client
)

func Init() {
	// Conexión a Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Conexión a MongoDB
	var err error
	MongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:example@localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Verificar la conexión a MongoDB
	if err := MongoClient.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to Redis and MongoDB")
}

func Close() {
	// Desconectar MongoDB
	if err := MongoClient.Disconnect(context.Background()); err != nil {
		log.Fatalf("Failed to disconnect MongoDB: %v", err)
	}
}
