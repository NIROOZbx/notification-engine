package app

import (
	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/db"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/handlers"
	"github.com/NIROOZbx/notification-engine/internal/middleware"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/session"

	"github.com/NIROOZbx/notification-engine/pkg/cache"
	"github.com/NIROOZbx/notification-engine/pkg/logger"
	"github.com/NIROOZbx/notification-engine/pkg/validator"
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
	Logger zerolog.Logger
}

type RouterDeps struct {
	App              *fiber.App
	AuthHandler      *handlers.AuthHandler
	UserHandler      *handlers.UserHandler
	WspHandler       *handlers.WorkspaceHandler
	AuthMiddleware   middleware.AuthMiddleware
	ApiKeyHandler    *handlers.APIKeyHandler
	ApiKeyMiddleware middleware.ApiKeyMiddleware
}


func StartApp(cfg *config.Config) (*App, error) {

	// ==========================================
	// 1. INFRASTRUCTURE & UTILS
	// ==========================================

	appLogger := logger.NewLogger(&cfg.Log)

	db, err := db.ConnectDB(&db.Config{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MinOpenConns:    cfg.Database.MinOpenConns,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxIdleTime:     cfg.Database.MaxIdleTime,
	})
	if err != nil {
		return nil, err
	}

	redis, err := cache.ConnectRedis(&cfg.Redis)

	if err != nil {
		return nil, err
	}

	v :=validator.NewValidator()

	// ==========================================
	// 2. REPOSITORIES & DATA STORES
	// ==========================================

	repo := sqlc.New(db)
	store := session.NewStore(redis)

	apiKeyRepo := repositories.NewAPIKeyRepository(repo)
	usrRepo := repositories.NewUserRepository(repo)
	wspRepo := repositories.NewWorkspaceRepository(repo,db)

	// ==========================================
	// 3. SERVICE LAYER (Business Logic)
	// ==========================================

	userService := services.NewUserService(usrRepo)
	workspaceService := services.NewWorkSpaceService(wspRepo)
	authService := services.NewAuthService(&cfg.Auth, userService, workspaceService, store)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)

	// ==========================================
	// 4. HTTP LAYER (Handlers & Middleware)
	// ==========================================

	userHandler := handlers.NewUserHandler(userService,appLogger)
	wspHandler := handlers.NewWorkspaceHandler(workspaceService)
	authHandler := handlers.NewAuthHandler(authService, &cfg.Auth, appLogger)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService, appLogger)

	// ==========================================
	// 5. FIBER SETUP & ROUTING
	// ==========================================
	app := fiber.New(fiber.Config{

		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		BodyLimit:    10 * 1024 * 1024,
		StructValidator: v,
	})
	

	authMiddleware := middleware.NewMiddleware(store, &cfg.Auth, appLogger,repo)
	apiKeyMiddleware := middleware.NewApiKeyMiddleware(apiKeyService, appLogger)

	r := RouterDeps{
		App:              app,
		AuthHandler:      authHandler,
		WspHandler:       wspHandler,
		UserHandler:      userHandler,
		AuthMiddleware:   authMiddleware,
		ApiKeyHandler:    apiKeyHandler,
		ApiKeyMiddleware: apiKeyMiddleware,
	}

	SetUpRoutes(&r)

	return &App{
		Server: app,
		Redis:  redis,
		DB:     db,
		Logger: appLogger,
	}, nil

}
