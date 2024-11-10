package routes

import (
	"exchange-rate/repository"
	"exchange-rate/utils/generate_transaction_code"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(app *fiber.App, mongoClient *mongo.Client, redisClient *redis.Client, codeGen *generate_transaction_code.CodeGenerator, transactionTypeRepo *repository.TransactionTypeRepository) {
	SetupAuthRoutes(app, mongoClient)
	SetupConversionRoutes(app, mongoClient, redisClient, codeGen)
	SetupTransactionTypeRoutes(app, transactionTypeRepo)
	SetupStatisticsRoutes(app, mongoClient)
	SetupTransactionsHistoryRoutes(app, mongoClient)
}
