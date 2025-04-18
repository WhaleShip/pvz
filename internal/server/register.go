package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/whaleship/pvz/internal/gen/oapi"
	"github.com/whaleship/pvz/internal/middleware"
)

func (srv *Server) RegisterAllHandlers(app *fiber.App) {
	wrapper := oapi.ServerInterfaceWrapper{
		Handler: srv,
	}

	srv.registerAuthHandlers(app, wrapper)
	srv.registerReceptionsHandlers(app, wrapper)
	srv.registerProductsHandlers(app, wrapper)
	srv.registerPvzHandlers(app, wrapper)
}

func (srv *Server) registerAuthHandlers(app *fiber.App, wrapper oapi.ServerInterfaceWrapper) {
	app.Post(
		"/dummyLogin",
		middleware.MetricsMiddleware("PostDummyLogin", srv.Metrics),
		wrapper.PostDummyLogin,
	)

	app.Post(
		"/login",
		middleware.MetricsMiddleware("PostLogin", srv.Metrics),
		wrapper.PostLogin,
	)

	app.Post(
		"/register",
		middleware.MetricsMiddleware("PostRegister", srv.Metrics),
		wrapper.PostRegister,
	)
}

func (srv *Server) registerPvzHandlers(app *fiber.App, wrapper oapi.ServerInterfaceWrapper) {
	app.Post(
		"/pvz",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("moderator"),
		middleware.MetricsMiddleware("PostPvz", srv.Metrics),
		wrapper.PostPvz,
	)

	app.Get(
		"/pvz",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee", "moderator"),
		middleware.MetricsMiddleware("GetPvz", srv.Metrics),
		wrapper.GetPvz,
	)
}

func (srv *Server) registerProductsHandlers(app *fiber.App, wrapper oapi.ServerInterfaceWrapper) {
	app.Post(
		"/products",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee"),
		middleware.MetricsMiddleware("PostProducts", srv.Metrics),
		wrapper.PostProducts,
	)

	app.Post(
		"/pvz/:pvzId/delete_last_product",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee"),
		middleware.MetricsMiddleware("PostPvzPvzIdDeleteLastProduct", srv.Metrics),
		wrapper.PostPvzPvzIdDeleteLastProduct,
	)
}

func (srv *Server) registerReceptionsHandlers(app *fiber.App, wrapper oapi.ServerInterfaceWrapper) {
	app.Post(
		"/receptions",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee"),
		middleware.MetricsMiddleware("PostReceptions", srv.Metrics),
		wrapper.PostReceptions,
	)

	app.Post(
		"/pvz/:pvzId/close_last_reception",
		middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee"),
		middleware.MetricsMiddleware("PostPvzPvzIdCloseLastReception", srv.Metrics),
		wrapper.PostPvzPvzIdCloseLastReception,
	)
}
