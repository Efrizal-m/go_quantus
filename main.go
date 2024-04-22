package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

func main() {
	// Connect to the database
	var err error

	err = godotenv.Load("config/.env")
	if err != nil {
		fmt.Println("failed load environtment")
	} else {
		fmt.Println("succes load environtment")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s "+
		"dbname=%s port=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGPORT"),
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the schema
	db.AutoMigrate(&User{})

	// Create a Fiber app
	app := fiber.New()

	// Routes
	app.Post("/register", register)
	app.Post("/login", login)

	// Protected routes
	protected := app.Group("/api")
	protected.Use(authenticate)
	protected.Get("/users", getUsers)
	protected.Post("/users", createUser)
	protected.Put("/users/:id", updateUser)
	protected.Delete("/users/:id", deleteUser)

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

func register(c *fiber.Ctx) error {
	// Parse request body
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	// Create user
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	return c.JSON(user)
}

func login(c *fiber.Ctx) error {
	// Parse request body
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}

	// Authenticate user (example: check username and password in the database)
	var foundUser User
	if err := db.Where("username = ?", user.Username).First(&foundUser).Error; err != nil {
		return fiber.ErrUnauthorized
	}

	// Check password
	if foundUser.Password != user.Password {
		return fiber.ErrUnauthorized
	}

	// Generate JWT token (example: using a JWT library)
	token := "exampleJWTToken"

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func authenticate(c *fiber.Ctx) error {
	// Example authentication middleware using JWT token
	// You should implement your own authentication logic here
	token := c.Get("Authorization")
	if token != "exampleJWTToken" {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

func getUsers(c *fiber.Ctx) error {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	return c.JSON(users)
}

func createUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return c.JSON(user)
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return err
	}
	if err := c.BodyParser(&user); err != nil {
		return err
	}
	db.Save(&user)
	return c.JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return err
	}
	db.Delete(&user)
	return nil
}
