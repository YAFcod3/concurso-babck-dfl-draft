package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionWithTypeInfo struct {
	TransactionID       primitive.ObjectID `bson:"_id,omitempty" json:"transactionId"`
	TransactionCode     string             `bson:"transaction_code" json:"transactionCode"`
	FromCurrency        string             `bson:"from_currency" json:"fromCurrency"`
	ToCurrency          string             `bson:"to_currency" json:"toCurrency"`
	Amount              float64            `bson:"amount" json:"amount"`
	AmountConverted     float64            `bson:"amount_converted" json:"amountConverted"`
	ExchangeRate        float64            `bson:"exchange_rate" json:"exchangeRate"`
	TransactionTypeID   primitive.ObjectID `bson:"transaction_type_id" json:"transactionTypeId"`
	TransactionTypeName string             `bson:"transactionTypeName" json:"transactionType"`
	CreatedAt           time.Time          `bson:"created_at" json:"createdAt"`
	UserID              string             `bson:"user_id" json:"userId"`
}

func GetFilteredTransactions(db *mongo.Database, filter bson.M) ([]TransactionWithTypeInfo, error) {
	collection := db.Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$lookup", bson.M{
			"from":         "transaction_types",
			"localField":   "transaction_type_id",
			"foreignField": "_id",
			"as":           "type_info",
		}}},
		{{"$unwind", bson.M{"path": "$type_info", "preserveNullAndEmptyArrays": true}}},
		{{"$project", bson.M{
			"transaction_code":    1,
			"from_currency":       1,
			"to_currency":         1,
			"amount":              1,
			"amount_converted":    1,
			"exchange_rate":       1,
			"transaction_type_id": 1,
			"created_at":          1,
			"user_id":             1,
			"transactionTypeName": "$type_info.name",
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []TransactionWithTypeInfo
	for cursor.Next(ctx) {
		var transaction TransactionWithTypeInfo
		if err := cursor.Decode(&transaction); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
