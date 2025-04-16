package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/server"
)

func main() {
	isPrefork := true
	app := fiber.New(fiber.Config{Prefork: isPrefork})
	app.Use(logger.New())
	var dbConn *pgxpool.Pool
	var err error

	if fiber.IsChild() || !isPrefork {
		dbConn, err = database.GetInitializedDB()
		if err != nil {
			log.Fatalf("db connection error: %v", err)
		}
		defer dbConn.Close()
	}

	server := gen.ServerInterface(server.NewServer(dbConn))

	gen.RegisterHandlers(fiber.Router(app), server)
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
