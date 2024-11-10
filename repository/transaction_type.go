package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionType struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description"`
}

type TransactionTypeRepository struct {
	collection *mongo.Collection
}

// Constructor para TransactionTypeRepository
func NewTransactionTypeRepository(db *mongo.Database) *TransactionTypeRepository {
	return &TransactionTypeRepository{
		collection: db.Collection("transaction_types"),
	}
}

func (r *TransactionTypeRepository) IsUniqueTransactionTypeName(name string) (bool, error) {
	filter := bson.M{"$expr": bson.M{
		"$eq": []interface{}{
			bson.M{"$toLower": name},
			bson.M{"$toLower": "$name"},
		},
	}}

	var existing TransactionType
	err := r.collection.FindOne(context.TODO(), filter).Decode(&existing)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true, nil // Nombre único
		}
		return false, err // Error en la consulta
	}

	return false, nil // Nombre ya existe
}

func (r *TransactionTypeRepository) Create(tt *TransactionType) error {
	_, err := r.collection.InsertOne(context.TODO(), tt)
	return err
}

func (r *TransactionTypeRepository) FindAll() ([]TransactionType, error) {
	var transactionTypes []TransactionType
	cursor, err := r.collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var tt TransactionType
		if err := cursor.Decode(&tt); err != nil {
			return nil, err
		}
		transactionTypes = append(transactionTypes, tt)
	}
	return transactionTypes, nil
}

func (r *TransactionTypeRepository) Update(id primitive.ObjectID, tt *TransactionType) error {
	_, err := r.collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": tt})
	return err
}

func (r *TransactionTypeRepository) Delete(id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
