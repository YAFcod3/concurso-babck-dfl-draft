package handlers

import (
	"exchange-rate/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTransactionType(c *fiber.Ctx, repo *repository.TransactionTypeRepository) error {
	var tt repository.TransactionType

	if err := c.BodyParser(&tt); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	isUnique, err := repo.IsUniqueTransactionTypeName(tt.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking uniqueness"})
	}
	if !isUnique {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Transaction type name must be unique"})
	}
	tt.ID = primitive.NewObjectID()
	err = repo.Create(&tt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction type"})
	}

	return c.Status(fiber.StatusCreated).JSON(tt)
}

func GetTransactionTypes(c *fiber.Ctx, repo *repository.TransactionTypeRepository) error {
	transactionTypes, err := repo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve transaction types"})
	}
	return c.JSON(transactionTypes)
}

func UpdateTransactionType(c *fiber.Ctx, repo *repository.TransactionTypeRepository) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var tt repository.TransactionType
	if err := c.BodyParser(&tt); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	isUnique, err := repo.IsUniqueTransactionTypeName(tt.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking uniqueness"})
	}
	if !isUnique {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Transaction type name must be unique"})
	}

	err = repo.Update(id, &tt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update transaction type"})
	}

	return c.JSON(tt)
}

func DeleteTransactionType(c *fiber.Ctx, repo *repository.TransactionTypeRepository) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	err = repo.Delete(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete transaction type"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
