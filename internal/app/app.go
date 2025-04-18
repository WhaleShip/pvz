package app

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/whaleship/pvz/internal/database"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
	"github.com/whaleship/pvz/internal/server"
	"google.golang.org/grpc"
)

type PVZApp struct {
	isPrefork bool
	pvz       *fiber.App
	srv       *server.Server
	grpcSrv   *grpc.Server
	db        database.PgxIface

	ipcManager *infrastructure.IPCManager
	aggregator *metrics.Aggregator
}

func New(isPrefork bool) *PVZApp {
	app := fiber.New(fiber.Config{Prefork: isPrefork})
	app.Use(logger.New())

	dbConn, err := database.GetInitializedDB()
	if err != nil {
		log.Fatalf("db connection error: %v", err)
	}
	return &PVZApp{
		isPrefork: isPrefork,
		pvz:       app,
		db:        &database.PgxPoolAdapter{Pool: dbConn},
	}
}

func (app *PVZApp) Start() {
	if app.aggregator != nil && app.ipcManager != nil {
		go app.startMetrics()
	}
	if app.pvz != nil {
		go app.startHTTPServer()
	}
	if app.grpcSrv != nil && !fiber.IsChild() {
		go app.startGRPCServer()
	}
}
