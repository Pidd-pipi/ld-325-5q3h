package router

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/handler"
	"github.com/blueship581/cybuildprice/backend/internal/middleware"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"log/slog"
)

func New(db *gorm.DB, logger *slog.Logger, jwtSecret string) *gin.Engine {
	v := validator.New()
	product := handler.NewProductHandler(service.NewProductService(repository.NewProductRepository(db), logger), v)
	offers := handler.NewOfferHandler(service.NewOfferService(repository.NewOfferRepository(db)), v)
	trend := handler.NewTrendHandler(service.NewPriceHistoryService(repository.NewPriceHistoryRepository(db)))
	user := handler.NewUserDataHandler(service.NewUserDataService(repository.NewUserDataRepository(db)), v)
	supplier := handler.NewSupplierHandler(service.NewSupplierService(repository.NewSupplierRepository(db)), v)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.RequestLogger(logger), middleware.ErrorHandler(), middleware.JWTOrDemoAuth(jwtSecret))
	r.GET(constants.HealthPath, func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api := r.Group(constants.APIPrefix)
	api.GET("/products", product.List)
	api.GET("/products/:id", product.Get)
	api.POST("/products/compare", product.Compare)
	api.GET("/products/:id/offers", offers.List)
	api.GET("/products/:id/trend", trend.Get)
	api.GET("/suppliers", supplier.List)
	api.PATCH("/admin/suppliers/:id/status", middleware.RequireRole(constants.RoleAdmin), supplier.UpdateStatus)
	api.PATCH("/supplier/offers/:id/status", middleware.RequireRole(constants.RoleSupplier, constants.RoleAdmin), offers.UpdateStatus)
	api.GET("/favorites", user.ListFavorites)
	api.POST("/favorites", user.CreateFavorite)
	api.POST("/alerts", user.CreateAlert)
	api.POST("/budgets", user.CreateBudget)
	return r
}
