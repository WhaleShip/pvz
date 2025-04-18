package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/whaleship/pvz/internal/app"
)

func main() {
	isPrefork := true
	pvzApp := app.New(isPrefork)

	pvzApp.InitializeMetrics()
	pvzApp.InitializeHTTPServer()
	pvzApp.InitializeGRPCServer()

	pvzApp.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}
