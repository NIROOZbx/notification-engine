package app

import (
	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/NIROOZbx/notification-engine/services/pkg/cache"
	"github.com/NIROOZbx/notification-engine/services/pkg/database"
	"github.com/NIROOZbx/notification-engine/services/pkg/logger"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type App struct {
	Server *fiber.App
	Redis  *redis.Client
	DB     *pgxpool.Pool
	Logger *zerolog.Logger
}

func StartApp(cfg *config.Config) (*App, error) {
	db, err := database.ConnectDB(&cfg.Database)
	if err != nil {
		return nil, err
	}
	app := fiber.New(fiber.Config{
		
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	})

	SetUpRoutes(app)

	redis, err := cache.ConnectRedis(&cfg.Redis)

	if err != nil {
		return nil, err
	}

	logger:=logger.NewLogger(&cfg.Log)

	return &App{
		Server: app,
		Redis:  redis,
		DB:     db,
		Logger: &logger,
	}, nil

}
