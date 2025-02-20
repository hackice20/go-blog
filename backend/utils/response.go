package utils

import (
	"backend/config"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func SuccessResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func GetDBCollection(collectionName string) *mongo.Collection {
	return config.DB.Collection(collectionName)
}

func UnauthorizedResponse(c *fiber.Ctx) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
}
