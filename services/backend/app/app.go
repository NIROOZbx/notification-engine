package app

import (
	"github.com/NIROOZbx/notification-engine/services/backend/config"
	authHdlrPkg "github.com/NIROOZbx/notification-engine/services/backend/internal/auth/handler"
	authSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/auth/services"
	sqlcDB "github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/middleware"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/session"
	userHdlrPkg "github.com/NIROOZbx/notification-engine/services/backend/internal/user/handler"
	userSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	wspHdlrPkg "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/handler"
	workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"
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

type RouterDeps struct {
	App         *fiber.App
	AuthHandler *authHdlrPkg.AuthHandler
	UserHandler *userHdlrPkg.UserHandler
	WspHandler  *wspHdlrPkg.WorkspaceHandler
	AuthMiddleware middleware.AuthMiddleware
}

func StartApp(cfg *config.Config) (*App, error) {

	// ==========================================
	// 1. INFRASTRUCTURE & UTILS
	// ==========================================

	appLogger := logger.NewLogger(&cfg.Log)

	db, err := database.ConnectDB(&cfg.Database)
	if err != nil {
		return nil, err
	}

	redis, err := cache.ConnectRedis(&cfg.Redis)

	if err != nil {
		return nil, err
	}

	// ==========================================
	// 2. REPOSITORIES & DATA STORES
	// ==========================================

	repo := sqlcDB.New(db)
	store := session.NewStore(redis)

	// ==========================================
	// 3. SERVICE LAYER (Business Logic)
	// ==========================================

	userService := userSvc.NewUserService(repo)
	workspaceService := workspaceSvc.NewService(repo, db)
	authService := authSvc.NewAuthService(&cfg.Auth, userService, workspaceService, store)

	// ==========================================
	// 4. HTTP LAYER (Handlers & Middleware)
	// ==========================================

	userHandler := userHdlrPkg.NewUserHandler(userService)
	wspHandler := wspHdlrPkg.NewWorkspaceHandler(workspaceService)
	authHandler := authHdlrPkg.NewAuthHandler(authService, &cfg.Auth)

	// ==========================================
	// 5. FIBER SETUP & ROUTING
	// ==========================================
	app := fiber.New(fiber.Config{

		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	})

	authMiddleware:=middleware.NewMiddleware(store,&cfg.Auth)

	r:=RouterDeps{
		App: app,
		AuthHandler: authHandler,
		WspHandler: wspHandler,
		UserHandler: userHandler,
		AuthMiddleware: authMiddleware,

	}



	SetUpRoutes(&r)

	return &App{
		Server: app,
		Redis:  redis,
		DB:     db,
		Logger: &appLogger,
	}, nil

}
