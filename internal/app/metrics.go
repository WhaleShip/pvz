package app

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/config"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/metrics"
)

func (app *PVZApp) InitializeMetrics() {
	techBuffer := 200
	businessBuffer := 400
	if app.isPrefork {
		techBuffer = 20
		businessBuffer = 40
	}

	aggregator := metrics.NewAggregator()
	ipcManager := infrastructure.NewIPCManager(config.IpcSockPath, techBuffer, businessBuffer, aggregator)

	app.aggregator = aggregator
	app.ipcManager = ipcManager
}

func (app *PVZApp) startMetrics() {
	if fiber.IsChild() || !app.isPrefork {
		app.ipcManager.StartSender()
	}

	if !fiber.IsChild() {
		app.ipcManager.StartServer()
		go func() {
			http.HandleFunc("/metrics", app.aggregator.HTTPHandler())
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
}
