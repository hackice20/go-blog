package routes

import (
	"backend/controllers"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Get("/user", controllers.User)
	auth.Post("/logout", controllers.Logout)

	// Blog routes
	blog := api.Group("/blogs")
	blog.Get("/", controllers.GetBlogs)
	blog.Get("/:id", controllers.GetBlog)
	blog.Post("/:id/like", JWTProtected(), controllers.LikeBlog)
	blog.Post("/:id/comment", JWTProtected(), controllers.CommentBlog)

	// Protected routes
	protected := blog.Group("", JWTProtected())
	protected.Post("/", controllers.CreateBlog)
	protected.Put("/:id", controllers.UpdateBlog)
	protected.Delete("/:id", controllers.DeleteBlog)
}

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract token from Cookie or Authorization header
		tokenString := c.Cookies("jwt")
		if tokenString == "" {
			// Check Authorization header
			authHeader := c.Get("Authorization")
			if len(authHeader) > 7 && strings.EqualFold(authHeader[0:7], "Bearer ") {
				tokenString = authHeader[7:]
			}
		}

		if tokenString == "" {
			fmt.Println("Unauthorized: No JWT token found in cookies or Authorization header")
			return controllers.UnauthorizedResponse(c)
		}

		// 2. Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(controllers.SecretKey), nil
		})

		if err != nil {
			fmt.Printf("JWT Validation Error: %v\n", err) // Detailed error
			return controllers.UnauthorizedResponse(c)
		}

		if !token.Valid {
			fmt.Println("Unauthorized: Invalid JWT token")
			return controllers.UnauthorizedResponse(c)
		}

		// 3. Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("Unauthorized: Failed to parse JWT claims")
			return controllers.UnauthorizedResponse(c)
		}

		// 4. Verify user ID existence in claims
		userID, exists := claims["id"].(string)
		if !exists || userID == "" {
			fmt.Println("Unauthorized: 'id' claim missing or invalid in JWT")
			return controllers.UnauthorizedResponse(c)
		}

		// 5. Validate userID format (optional)
		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			fmt.Printf("Invalid userID format in JWT: %v\n", userID)
			return controllers.UnauthorizedResponse(c)
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
