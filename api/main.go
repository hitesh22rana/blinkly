// Path: api\main.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/hitesh22rana/blinkly/routes"
)

var PORT string = os.Getenv("APP_PORT")

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	app := fiber.New(fiber.Config{
		AppName: "Blinkly",
	})

	app.Use(logger.New())

	setupRoutes(app)

	log.Fatal(app.Listen(PORT))
}
