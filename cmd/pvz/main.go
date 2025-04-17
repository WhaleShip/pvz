package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whaleship/pvz/internal/config"
	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
	"github.com/whaleship/pvz/internal/server"
)

func main() {
	isPrefork := true
	app := fiber.New(fiber.Config{Prefork: isPrefork})
	app.Use(logger.New())
	var dbConn *pgxpool.Pool
	var err error

	techBuffer := 200
	businessBuffer := 400
	if isPrefork {
		techBuffer = 20
		businessBuffer = 40
	}

	aggregator := metrics.NewAggregator()
	ipcManager := infrastructure.NewIPCManager(config.IpcSockPath, techBuffer, businessBuffer, aggregator)

	if fiber.IsChild() || !isPrefork {
		dbConn, err = database.GetInitializedDB()
		if err != nil {
			log.Fatalf("db connection error: %v", err)
		}
		defer dbConn.Close()

		ipcManager.StartSender()
	}

	if !fiber.IsChild() {
		ipcManager.StartServer()
		go func() {
			http.HandleFunc("/metrics", aggregator.HTTPHandler())
			srv := &http.Server{
				Addr:         ":9000",
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
			}
			log.Println("metrics are available on :9000/metrics")
			if err := srv.ListenAndServe(); err != nil {
				log.Fatalf("metrics server error: %v", err)
			}
		}()
	}

	srv := server.NewServer(dbConn, ipcManager)
	srv.RegisterAllHandlers(app)

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
