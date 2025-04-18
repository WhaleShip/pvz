package app

import (
	"log"
	"net"
)

func (app *PVZApp) InitializeGRPCServer() {
	srv := app.srv.RegisterGRPCHandlers()
	app.grpcSrv = srv
}

func (app *PVZApp) startGRPCServer() {
	// #nosec G102
	lis, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		log.Fatalf("failed to listen on port 3000: %v", err)
	}

	log.Println("gRPC server starting on :3000")
	if err := app.grpcSrv.Serve(lis); err != nil {
		log.Printf("gRPC server error: %v", err)
	}
}
