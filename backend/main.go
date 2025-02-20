package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"backend/config"
	"backend/routes"
)

func main() {
	// Connect to MongoDB
	config.ConnectDB()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	// Setup routes
	routes.SetupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}
