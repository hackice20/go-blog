package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"backend/models"
	"backend/utils"
)

func CreateBlog(c *fiber.Ctx) error {
	var data struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&data); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request")
	}

	userID := c.Locals("userID").(string)
	objID, _ := primitive.ObjectIDFromHex(userID)

	blog := models.Blog{
		Title:     data.Title,
		Content:   data.Content,
		Author:    objID,
		Likes:     []primitive.ObjectID{},
		Comments:  []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := utils.GetDBCollection("blogs")
	result, err := collection.InsertOne(context.TODO(), blog)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create blog")
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Blog created", result)
}

func GetBlogs(c *fiber.Ctx) error {
	collection := utils.GetDBCollection("blogs")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch blogs")
	}

	var blogs []models.Blog
	if err = cursor.All(context.TODO(), &blogs); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to decode blogs")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blogs retrieved", blogs)
}

func GetBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID")
	}

	collection := utils.GetDBCollection("blogs")
	var blog models.Blog
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&blog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Blog not found")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch blog")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog retrieved", blog)
}

func UpdateBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID")
	}

	var data struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&data); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request")
	}

	userID := c.Locals("userID").(string)
	authorID, _ := primitive.ObjectIDFromHex(userID)

	collection := utils.GetDBCollection("blogs")
	result := collection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": objID, "author": authorID},
		bson.M{"$set": bson.M{
			"title":      data.Title,
			"content":    data.Content,
			"updated_at": time.Now(),
		}},
	)

	if result.Err() != nil {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Unauthorized or blog not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog updated", nil)
}

func DeleteBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID")
	}

	userID := c.Locals("userID").(string)
	authorID, _ := primitive.ObjectIDFromHex(userID)

	collection := utils.GetDBCollection("blogs")
	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID, "author": authorID})
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete blog")
	}

	if result.DeletedCount == 0 {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Unauthorized or blog not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog deleted", nil)
}

func LikeBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Error: Invalid blog ID format") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid blog ID")
	}

	// Extract userID safely
	user, ok := c.Locals("userID").(string)
	if !ok || user == "" {
		fmt.Println("Unauthorized: Missing or invalid userID in Locals") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	userObjID, err := primitive.ObjectIDFromHex(user)
	if err != nil {
		fmt.Println("Error: Invalid user ID format") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format")
	}

	collection := utils.GetDBCollection("blogs")
	result, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$addToSet": bson.M{"likes": userObjID}},
	)

	if err != nil {
		fmt.Println("Error: Failed to like blog ->", err) // Debugging log
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to like blog")
	}

	if result.ModifiedCount == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Blog not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog liked", nil)
}

func CommentBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Error: Invalid blog ID format") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid blog ID")
	}

	// Parse request body safely
	var data struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&data); err != nil || strings.TrimSpace(data.Text) == "" {
		fmt.Println("Error: Invalid or empty comment text") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid comment text")
	}

	// Extract userID safely
	user, ok := c.Locals("userID").(string)
	if !ok || user == "" {
		fmt.Println("Unauthorized: Missing or invalid userID in Locals") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	userObjID, err := primitive.ObjectIDFromHex(user)
	if err != nil {
		fmt.Println("Error: Invalid user ID format") // Debugging log
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID format")
	}

	comment := models.Comment{
		Text:      data.Text,
		Author:    userObjID,
		CreatedAt: time.Now(),
	}

	collection := utils.GetDBCollection("blogs")
	result, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$push": bson.M{"comments": comment}},
	)

	if err != nil {
		fmt.Println("Error: Failed to add comment ->", err) // Debugging log
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add comment")
	}

	if result.ModifiedCount == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Blog not found")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Comment added", comment)
}
