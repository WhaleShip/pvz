package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/handlers"
	"github.com/whaleship/pvz/internal/service"
)

func main() {
	isPrefork := true
	app := fiber.New(fiber.Config{Prefork: isPrefork})
	app.Use(logger.New())

	if fiber.IsChild() || !isPrefork {
		dbConn, err := database.GetInitializedDB()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("db", dbConn)
			return c.Next()
		})
		if err != nil {
			log.Fatalf("db connection error: %v", err)
		}
		defer dbConn.Close()
	}

	authSvc := service.NewAuthService()
	authHandler := handlers.NewAuthHandler(authSvc)
	app.Post("/dummyLogin", authHandler.PostDummyLogin)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatalln(err)
		}
	}()

	<-quit

	if err := app.Shutdown(); err != nil {
		log.Printf("fiber shutdown error: %v", err)
	}
}
