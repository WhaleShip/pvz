package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/handlers"
	"github.com/whaleship/pvz/internal/middleware"
	"github.com/whaleship/pvz/internal/repository"
	"github.com/whaleship/pvz/internal/service"
)

type Server struct {
	AuthHandler      *handlers.AuthHandler
	PVZHandler       *handlers.PVZHandler
	ProductHandler   *handlers.ProductHandler
	ReceptionHandler *handlers.ReceptionHandler
}

func (srv *Server) PostDummyLogin(c *fiber.Ctx) error {
	return srv.AuthHandler.PostDummyLogin(c)
}

func (srv *Server) PostLogin(c *fiber.Ctx) error {
	return srv.AuthHandler.PostLogin(c)
}

func (srv *Server) PostRegister(c *fiber.Ctx) error {
	return srv.AuthHandler.PostRegister(c)
}

func (srv *Server) PostPvz(c *fiber.Ctx) error {
	return srv.PVZHandler.PostPvz(c)
}

func (srv *Server) GetPvz(c *fiber.Ctx, params gen.GetPvzParams) error {
	return srv.PVZHandler.GetPvz(c)
}

func (srv *Server) PostProducts(c *fiber.Ctx) error {
	return srv.ProductHandler.PostProducts(c)
}

func (srv *Server) PostPvzPvzIdDeleteLastProduct(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	return srv.ProductHandler.PostPvzPvzIdDeleteLastProduct(c, pvzId)
}

func (srv *Server) PostPvzPvzIdCloseLastReception(c *fiber.Ctx, pvzId openapi_types.UUID) error {
	return srv.ReceptionHandler.CloseReception(c, pvzId)
}

func (srv *Server) PostReceptions(c *fiber.Ctx) error {
	return srv.ReceptionHandler.PostReception(c)
}

func NewServer(conn *pgxpool.Pool) *Server {
	userRepo := repository.NewUserRepository(conn)
	pvzRepo := repository.NewPVZRepository(conn)
	productRepo := repository.NewProductRepository(conn)
	receptionRepo := repository.NewReceptionRepository(conn)

	authSvc := service.NewAuthService(userRepo)
	pvzSvc := service.NewPVZService(pvzRepo)
	productSvc := service.NewProductService(productRepo)
	receptionSvc := service.NewReceptionService(receptionRepo)

	authHandler := handlers.NewAuthHandler(authSvc)
	pvzHandler := handlers.NewPVZHandler(pvzSvc)
	productHandler := handlers.NewProductHandler(productSvc)
	receptionHandler := handlers.NewReceptionHandler(receptionSvc)

	return &Server{
		AuthHandler:      authHandler,
		PVZHandler:       pvzHandler,
		ProductHandler:   productHandler,
		ReceptionHandler: receptionHandler,
	}
}

func (srv *Server) RegisterAllHandlers(app *fiber.App) {
	wrapper := gen.ServerInterfaceWrapper{Handler: srv}
	app.Post("/dummyLogin", wrapper.PostDummyLogin)
	app.Post("/login", wrapper.PostLogin)
	app.Post("/register", wrapper.PostRegister)
	app.Post("/pvz", middleware.AuthMiddleware, middleware.RoleMiddleware("moderator"), wrapper.PostPvz)
	app.Post("/products", middleware.AuthMiddleware, middleware.RoleMiddleware("employee"), wrapper.PostProducts)
	app.Post("/receptions", middleware.AuthMiddleware, middleware.RoleMiddleware("employee"), wrapper.PostReceptions)
	app.Post("/pvz/:pvzId/close_last_reception", middleware.AuthMiddleware,
		middleware.RoleMiddleware("employee"), wrapper.PostPvzPvzIdCloseLastReception)
	app.Post("/pvz/:pvzId/delete_last_product",
		middleware.AuthMiddleware, middleware.RoleMiddleware("employee"), wrapper.PostPvzPvzIdDeleteLastProduct)
	app.Get("/pvz", middleware.AuthMiddleware, middleware.RoleMiddleware("employee", "moderator"), wrapper.GetPvz)
}
