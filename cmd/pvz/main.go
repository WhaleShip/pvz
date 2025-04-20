package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/whaleship/pvz/internal/app"
)

func main() {
	isPrefork, err := GetIsPrefork()
	if err != nil {
		log.Fatalf("prefork env var not set")
	}

	pvzApp := app.New(isPrefork)

	pvzApp.InitDBConnection()
	pvzApp.InitializeMetrics()
	pvzApp.InitializeHTTPServer()
	pvzApp.InitializeGRPCServer()

	pvzApp.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}

func GetIsPrefork() (bool, error) {
	isPreforkStr, _ := os.LookupEnv("IS_PREFORK")
	res, err := strconv.ParseBool(isPreforkStr)
	if err != nil {
		return false, err
	}
	return res, nil
}
