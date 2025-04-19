package app

import (
	"log"

	"github.com/whaleship/pvz/internal/server"
)

func (app *PVZApp) InitializeHTTPServer() {
	srv := server.NewServer(app.db, app.ipcManager)
	srv.RegisterHttpHandlers(app.PVZ)
	app.srv = srv
}
func (app *PVZApp) startHTTPServer() {
	if err := app.PVZ.Listen(":8080"); err != nil {
		log.Fatalln(err)
	}
}
