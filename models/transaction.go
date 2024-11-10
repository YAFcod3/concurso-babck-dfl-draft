package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionStatus string

type Transaction struct {
	TransactionCode   string             `bson:"transaction_code" json:"transactionCode"`
	FromCurrency      string             `bson:"from_currency" json:"fromCurrency"`
	ToCurrency        string             `bson:"to_currency" json:"toCurrency"`
	Amount            float64            `bson:"amount" json:"amount"`
	AmountConverted   float64            `bson:"amount_converted" json:"amountConverted"`
	ExchangeRate      float64            `bson:"exchange_rate" json:"exchangeRate"`
	TransactionTypeID primitive.ObjectID `bson:"transaction_type_id" json:"transactionTypeId"`
	CreatedAt         time.Time          `bson:"created_at" json:"createdAt"`
	UserID            string             `bson:"user_id" json:"userId"`
	Status            TransactionStatus  `bson:"status"`
}
