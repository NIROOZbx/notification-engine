package app

import (
	"context"
	"net/http"
	"sync"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/consts"
	"github.com/NIROOZbx/notification-engine/db"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/NIROOZbx/notification-engine/engine/notification/queue"
	"github.com/NIROOZbx/notification-engine/engine/notification/scheduler"
	"github.com/NIROOZbx/notification-engine/engine/notification/sender/email"
	"github.com/NIROOZbx/notification-engine/engine/notification/sender/sms"
	"github.com/NIROOZbx/notification-engine/engine/notification/template"
	"github.com/NIROOZbx/notification-engine/internal/billing"
	"github.com/NIROOZbx/notification-engine/internal/handlers"
	"github.com/NIROOZbx/notification-engine/internal/metrics"
	"github.com/NIROOZbx/notification-engine/internal/middleware"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/session"

	"github.com/NIROOZbx/notification-engine/pkg/cache"
	"github.com/NIROOZbx/notification-engine/pkg/httpclient"
	"github.com/NIROOZbx/notification-engine/pkg/logger"
	"github.com/NIROOZbx/notification-engine/pkg/serializer"
	"github.com/NIROOZbx/notification-engine/pkg/validator"

	"fmt"

	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/encryptor"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type App struct {
	Server    *fiber.App
	Redis     *redis.Client
	DB        *pgxpool.Pool
	Logger    zerolog.Logger
	Consumer  map[string]queue.Consumer
	Engine    *core.Engine
	Scheduler *scheduler.Scheduler
	wg        *sync.WaitGroup
	Producer  core.Producer
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
	PlanHandler       *handlers.PlanHandler
	BillingHandler    *handlers.BillingHandler
	AnalyticsHandler  *handlers.AnalyticsHandler
	Logger            zerolog.Logger
	Metrics           *metrics.Metrics
}

func StartApp(cfg *config.Config) (*App, error) {

	// ==========================================
	// 1. INFRASTRUCTURE & UTILS
	// ==========================================

	appLogger := logger.NewLogger(&cfg.Log, cfg.Auth.Environment)

	prom := metrics.NewMetrics()

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

	httpClient := httpclient.NewClient()

	billingClient, err := billing.NewGRPCClient(cfg.GRPC.GRPCAddr, appLogger)

	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to connect to billing service")
	}
	// ==========================================
	// 2. REPOSITORIES & DATA STORES
	// ==========================================

	repo := sqlc.New(db)
	store := session.NewStore(redis)

	apiKeyRepo := repositories.NewAPIKeyRepository(repo)
	usrRepo := repositories.NewUserRepository(repo)
	wspRepo := repositories.NewWorkspaceRepository(repo, db)
	chnlConfigRepo := repositories.NewChannelConfigRepo(repo, db)
	planRepo := repositories.NewPlanRepository(repo)
	subscriberRepo := repositories.NewSubscriberRepo(repo)
	templateRepo := repositories.NewTemplateRepository(repo)
	notifRepo := repositories.NewNotificationRepository(repo, chnlConfigRepo, templateRepo)
	layoutRepo := repositories.NewLayoutRepo(repo)
	schedulerRepo := repositories.NewSchedulerRepo(repo)
	analyticsRepo := repositories.NewAnalyticsRepository(repo)

	// ==========================================
	// 3. SERVICE LAYER (Business Logic)
	// ==========================================

	userService := services.NewUserService(usrRepo)
	workspaceService := services.NewWorkSpaceService(wspRepo, billingClient)
	authService := services.NewAuthService(&cfg.Auth, userService, workspaceService, store)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo, appLogger)
	subscriberSvc := services.NewSubscriberService(subscriberRepo)
	chnlConfigSvc := services.NewChannelConfigService(chnlConfigRepo, cfg.SecretKey)
	templateSvc := services.NewTemplateService(templateRepo, layoutRepo, wspRepo, chnlConfigRepo)
	layoutSvc := services.NewLayoutService(layoutRepo, wspRepo)
	planSvc := services.NewPlanService(planRepo)
	billingSvc := services.NewBillingService(billingClient)
	analyticsSvc := services.NewAnalyticsService(analyticsRepo, appLogger)

	// ==========================================
	//  ENGINE CONFIGURATION
	// ==========================================

	producer := queue.NewProducer(kafkaCfg.Broker, prom)

	render := template.NewRenderer()

	engine := core.NewEngine(core.EngineConfig{
		Repo:          notifRepo,
		Producer:      producer,
		Log:           appLogger,
		Renderer:      render,
		SecretKey:     cfg.SecretKey,
		BillingClient: billingClient,
		Metrics:       prom,
	})

	setUpMockProviders(engine, appLogger)

	s := scheduler.NewScheduler(producer, appLogger, schedulerRepo, consts.Interval)

	setUpProviders(engine, appLogger, httpClient)

	consumers := setUpConsumers(kafkaCfg.Broker, engine, kafkaCfg.GroupID, appLogger, prom)

	// ==========================================
	// 4. HTTP LAYER (Handlers & Middleware)
	// ==========================================

	userHandler := handlers.NewUserHandler(userService, workspaceService, appLogger)
	wspHandler := handlers.NewWorkspaceHandler(workspaceService)
	authHandler := handlers.NewAuthHandler(authService, &cfg.Auth, appLogger, store, engine)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService, appLogger)
	notifHandler := handlers.NewNotificationHandler(engine, notifRepo, appLogger)
	subscriberHandler := handlers.NewSubscriberHandler(subscriberSvc, appLogger)
	templateHandler := handlers.NewTemplateHandler(templateSvc, appLogger)
	layoutHandler := handlers.NewLayoutHandler(layoutSvc, appLogger)
	chnlConfigHandler := handlers.NewChannelConfigHandler(chnlConfigSvc, appLogger)
	planHandler := handlers.NewPlanHandler(planSvc)
	billingHandler := handlers.NewBillingHandler(billingSvc, userService, appLogger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsSvc, appLogger)

	// ==========================================
	// 5. FIBER SETUP & ROUTING
	// ==========================================
	app := fiber.New(fiber.Config{

		JSONEncoder:     serializer.Marshal,
		JSONDecoder:     serializer.Unmarshal,
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
		PlanHandler:       planHandler,
		BillingHandler:    billingHandler,
		AnalyticsHandler:  analyticsHandler,
		Logger:            appLogger,
		Metrics:           prom,
	}

	SetUpRoutes(&r, &cfg.CORS)

	a := &App{
		Server:    app,
		Redis:     redis,
		DB:        db,
		Logger:    appLogger,
		Consumer:  consumers,
		Engine:    engine,
		Scheduler: s,
		Producer:  producer,
		wg:        &sync.WaitGroup{},
	}

	if err := a.BootstrapSystem(context.Background(), cfg, chnlConfigRepo); err != nil {
		appLogger.Error().Err(err).Msg("failed to bootstrap system workspace")
	}

	return a, nil
}

func setUpProviders(e *core.Engine, log zerolog.Logger, httpClient *http.Client) {
	e.RegisterProvider(email.NewSendGridProvider(log, httpClient))
	e.RegisterProvider(email.NewSESProvider(log, httpClient))
	e.RegisterProvider(email.NewResendProvider(log, httpClient))
	e.RegisterProvider(sms.NewTwilioProvider(log, httpClient))
}

func setUpMockProviders(e *core.Engine, log zerolog.Logger) {

	channels := []string{"email", "sms", "push"}

	for _, val := range channels {
		mockProvider := provider.NewMockProvider(val, log)
		e.RegisterMockProvider(mockProvider)
	}

}

func setUpConsumers(broker string, engine *core.Engine, groupID string, log zerolog.Logger, metrics *metrics.Metrics) map[string]queue.Consumer {

	consumers := make(map[string]queue.Consumer)

	topics := []string{queue.TopicSMS, queue.TopicEmail, queue.TopicDLQ, queue.TopicSystem}

	for _, topic := range topics {
		handler := engine.Process
		if topic == queue.TopicDLQ {
			handler = engine.ProcessDLQ
		}

		taggedLogger := log.With().Str("worker_topic", topic).Logger()
		consumers[topic] = queue.NewConsumer(broker, topic, groupID, handler, taggedLogger, metrics)
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

func (a *App) StartScheduler(ctx context.Context) {
	a.Logger.Info().Msg("background scheduler started")
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.Scheduler.Start(ctx)
	}()
}

func (a *App) StopScheduler() {
	a.Logger.Info().Msg("waiting for scheduler to finish its current batch...")
	a.wg.Wait()
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

func (a *App) BootstrapSystem(ctx context.Context, cfg *config.Config, chnlRepo repositories.ChannelConfigRepo) error {
	wsID, err := utils.StringToUUID(cfg.Auth.SystemWorkspaceID)
	if err != nil {
		return err
	}

	existing, err := chnlRepo.GetDefaultChannelConfig(ctx, wsID, "email")
	if err == nil && existing != nil {
		a.Logger.Info().Msg("system email configuration already exists")
		return nil
	}

	a.Logger.Info().Msg("bootstrapping system email configuration...")

	creds := map[string]string{
		"api_key":    cfg.Auth.SystemEmailAPIKey,
		"from_email": cfg.Auth.SystemFromEmail,
	}

	encrypted, err := encryptor.EncryptMap(creds, cfg.SecretKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt system credentials: %w", err)
	}

	provider := cfg.Auth.SystemEmailProvider

	params := domain.CreateChannelConfigParams{
		WorkspaceID: wsID,
		Channel:     consts.ChannelEmail,
		Provider:    provider,
		DisplayName: "System Email Provider",
		Credentials: creds,
		IsActive:    true,
		IsDefault:   true,
	}

	_, err = chnlRepo.Create(ctx, encrypted, params)
	if err != nil {
		return fmt.Errorf("failed to create system channel config: %w", err)
	}

	a.Logger.Info().Msg("system email configuration bootstrapped successfully")
	return nil
}
