package app

import (
	"context"
	"sync"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/db"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/NIROOZbx/notification-engine/engine/notification/queue"
	"github.com/NIROOZbx/notification-engine/engine/notification/template"
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
	Server   *fiber.App
	Redis    *redis.Client
	DB       *pgxpool.Pool
	Logger   zerolog.Logger
	Producer core.Producer
	Consumer map[string]queue.Consumer
	Engine   *core.Engine
	wg       *sync.WaitGroup
}

type RouterDeps struct {
	App               *fiber.App
	AuthHandler       *handlers.AuthHandler
	UserHandler       *handlers.UserHandler
	WspHandler        *handlers.WorkspaceHandler
	AuthMiddleware    middleware.AuthMiddleware
	ApiKeyHandler     *handlers.APIKeyHandler
	ApiKeyMiddleware  middleware.ApiKeyMiddleware
	NotifHandler      *handlers.NotificationHandler
	SubscriberHandler *handlers.SubscriberHandler
	TemplateHandler   *handlers.TemplateHandler
	LayoutHandler     *handlers.LayoutHandler
	ChnlConfigHandler *handlers.ChannelConfigHandler
}

func StartApp(cfg *config.Config) (*App, error) {

	// ==========================================
	// 1. INFRASTRUCTURE & UTILS
	// ==========================================

	appLogger := logger.NewLogger(&cfg.Log)

	kafkaCfg := cfg.Kafka

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

	v := validator.NewValidator()

	// ==========================================
	// 2. REPOSITORIES & DATA STORES
	// ==========================================

	repo := sqlc.New(db)
	store := session.NewStore(redis)

	apiKeyRepo := repositories.NewAPIKeyRepository(repo)
	usrRepo := repositories.NewUserRepository(repo)
	wspRepo := repositories.NewWorkspaceRepository(repo, db)
	notifRepo := repositories.NewNotificationRepository(repo)
	subscriberRepo := repositories.NewSubscriberRepo(repo)
	templateRepo := repositories.NewTemplateRepository(repo)
	layoutRepo := repositories.NewLayoutRepo(repo)
	chnlConfigRepo := repositories.NewChannelConfigRepo(repo)

	// ==========================================
	// 3. SERVICE LAYER (Business Logic)
	// ==========================================

	userService := services.NewUserService(usrRepo)
	workspaceService := services.NewWorkSpaceService(wspRepo)
	authService := services.NewAuthService(&cfg.Auth, userService, workspaceService, store)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	subscriberSvc := services.NewSubscriberService(subscriberRepo)
	templateSvc := services.NewTemplateService(templateRepo)
	layoutSvc := services.NewLayoutService(layoutRepo)
	chnlConfigSvc := services.NewChannelConfigService(chnlConfigRepo, cfg.SecretKey)

	// ==========================================
	//  ENGINE CONFIGURATION
	// ==========================================

	producer := queue.NewProducer(kafkaCfg.Broker)

	render := template.NewRenderer()

	engine := core.NewEngine(notifRepo, producer, appLogger, render)

	setUpProviders(engine,appLogger)

	consumers := setUpConsumers(kafkaCfg.Broker, engine, kafkaCfg.GroupID, appLogger)

	// ==========================================
	// 4. HTTP LAYER (Handlers & Middleware)
	// ==========================================

	userHandler := handlers.NewUserHandler(userService, appLogger)
	wspHandler := handlers.NewWorkspaceHandler(workspaceService)
	authHandler := handlers.NewAuthHandler(authService, &cfg.Auth, appLogger)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService, appLogger)
	notifHandler := handlers.NewNotificationHandler(engine, notifRepo, appLogger)
	subscriberHandler := handlers.NewSubscriberHandler(subscriberSvc, appLogger)
	templateHandler := handlers.NewTemplateHandler(templateSvc, appLogger)
	layoutHandler := handlers.NewLayoutHandler(layoutSvc, appLogger)
	chnlConfigHandler := handlers.NewChannelConfigHandler(chnlConfigSvc, appLogger)

	// ==========================================
	// 5. FIBER SETUP & ROUTING
	// ==========================================
	app := fiber.New(fiber.Config{

		JSONEncoder:     sonic.Marshal,
		JSONDecoder:     sonic.Unmarshal,
		IdleTimeout:     cfg.Server.IdleTimeout,
		ReadTimeout:     cfg.Server.ReadTimeout,
		WriteTimeout:    cfg.Server.WriteTimeout,
		BodyLimit:       10 * 1024 * 1024,
		StructValidator: v,
	})

	authMiddleware := middleware.NewMiddleware(store, &cfg.Auth, appLogger, repo)
	apiKeyMiddleware := middleware.NewApiKeyMiddleware(apiKeyService, appLogger)

	r := RouterDeps{
		App:               app,
		AuthHandler:       authHandler,
		WspHandler:        wspHandler,
		UserHandler:       userHandler,
		AuthMiddleware:    authMiddleware,
		ApiKeyHandler:     apiKeyHandler,
		ApiKeyMiddleware:  apiKeyMiddleware,
		NotifHandler:      notifHandler,
		SubscriberHandler: subscriberHandler,
		TemplateHandler:   templateHandler,
		LayoutHandler:     layoutHandler,
		ChnlConfigHandler: chnlConfigHandler,
	}

	SetUpRoutes(&r)

	return &App{
		Server:   app,
		Redis:    redis,
		DB:       db,
		Logger:   appLogger,
		Producer: producer,
		Consumer: consumers,
		Engine:   engine,
		wg:       &sync.WaitGroup{},
	}, nil

}

func setUpProviders(e *core.Engine,log zerolog.Logger) {

	channels := []string{"email", "sms", "push"}

	for _, val := range channels {
		mockProvider := provider.NewMockProvider(val,log)

		e.RegisterMockProvider(mockProvider)
	}

}

func setUpConsumers(broker string, engine *core.Engine, groupID string, log zerolog.Logger) map[string]queue.Consumer {

	consumers := make(map[string]queue.Consumer)

	topics := []string{queue.TopicDLQ, queue.TopicEmail, queue.TopicInApp, queue.TopicRetry, queue.TopicDLQ}

	for _, topic := range topics {
		workerName := topic

		taggedLogger := log.With().Str("worker_topic", workerName).Logger()
		consumers[topic] = queue.NewConsumer(broker, topic, groupID, engine.Process, taggedLogger)
	}

	return consumers
}

func (a *App) StartConsumers(ctx context.Context) {
	

	for topic, c := range a.Consumer {
		a.Logger.Info().Str("topic", topic).Msg("consumer started")
		a.wg.Add(1)
		go func() {

			defer a.wg.Done()
			c.Start(ctx)
		}()
	}
}

func (a *App) StopConsumers() {
	a.Logger.Info().Msg("waiting for consumers to finish their current tasks...")
	a.wg.Wait()

	for topic, c := range a.Consumer {
		a.Logger.Info().Str("topic", topic).Msg("closing kafka consumer connection...")
		if err := c.Close(); err != nil {
			a.Logger.Error().Err(err).Str("topic", topic).Msg("failed to delicately close consumer")
		}
	}
}
