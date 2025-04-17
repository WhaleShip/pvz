package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/whaleship/pvz/internal/gen"
	"github.com/whaleship/pvz/internal/handlers"
	"github.com/whaleship/pvz/internal/infrastructure"
	"github.com/whaleship/pvz/internal/repository"
	"github.com/whaleship/pvz/internal/service"
)

type Server struct {
	AuthHandler      *handlers.AuthHandler
	PVZHandler       *handlers.PVZHandler
	ProductHandler   *handlers.ProductHandler
	ReceptionHandler *handlers.ReceptionHandler
	Metrics          *infrastructure.IPCManager
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

func NewServer(conn *pgxpool.Pool, ipcManager *infrastructure.IPCManager) *Server {
	userRepo := repository.NewUserRepository(conn)
	pvzRepo := repository.NewPVZRepository(conn)
	productRepo := repository.NewProductRepository(conn)
	receptionRepo := repository.NewReceptionRepository(conn)

	authSvc := service.NewAuthService(userRepo)
	pvzSvc := service.NewPVZService(pvzRepo, receptionRepo, productRepo, ipcManager)
	productSvc := service.NewProductService(productRepo, ipcManager)
	receptionSvc := service.NewReceptionService(receptionRepo, ipcManager)

	authHandler := handlers.NewAuthHandler(authSvc)
	pvzHandler := handlers.NewPVZHandler(pvzSvc)
	productHandler := handlers.NewProductHandler(productSvc)
	receptionHandler := handlers.NewReceptionHandler(receptionSvc)

	return &Server{
		AuthHandler:      authHandler,
		PVZHandler:       pvzHandler,
		ProductHandler:   productHandler,
		ReceptionHandler: receptionHandler,
		Metrics:          ipcManager,
	}
}
