package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"backend/models"
	"backend/utils"
)

var SecretKey = "secretkey" // Change this in production

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request")
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Username:  data["username"],
		Email:     data["email"],
		Password:  string(password),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := utils.GetDBCollection("users")
	_, err := collection.InsertOne(context.TODO(), user)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Email already exists")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Registration failed")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User registered", nil)
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request")
	}

	var user models.User
	collection := utils.GetDBCollection("users")
	err := collection.FindOne(context.TODO(), bson.M{"email": data["email"]}).Decode(&user)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"])); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid credentials")
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Login failed")
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", fiber.Map{"token": token})
}

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthenticated")
	}

	claims := token.Claims.(jwt.MapClaims)

	var user models.User
	collection := utils.GetDBCollection("users")
	objID, _ := primitive.ObjectIDFromHex(claims["id"].(string))
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User found", user)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return utils.SuccessResponse(c, fiber.StatusOK, "Logout successful", nil)
}



func UnauthorizedResponse(c *fiber.Ctx) error {

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{

		"error": "Unauthorized",
	})

}
