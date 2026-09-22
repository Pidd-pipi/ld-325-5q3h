package main

import (
	"context"
	"fmt"
	"os"

	"github.com/blueship581/cybuildprice/backend/internal/config"
	"github.com/blueship581/cybuildprice/backend/internal/logger"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	log := logger.New()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{})
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	if err := model.Migrate(db); err != nil {
		log.Error("migrate database", "error", err)
		os.Exit(1)
	}
	if err := model.Seed(db); err != nil {
		log.Error("seed database", "error", err)
		os.Exit(1)
	}
	cache := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress()})
	defer func() {
		if closeErr := cache.Close(); closeErr != nil {
			log.Warn("close redis", "error", closeErr)
		}
	}()
	if err := cache.Ping(context.Background()).Err(); err != nil {
		log.Warn("redis unavailable; continuing without cache", "error", err)
	}
	app := router.New(db, log, cfg.JWTSecret)
	log.Info("server starting", "address", cfg.ServerAddress())
	if err := app.Run(cfg.ServerAddress()); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
