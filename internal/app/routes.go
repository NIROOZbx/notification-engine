package app

import (
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetUpRoutes(r *RouterDeps) {
	r.App.Use(recover.New())
	r.App.Use(logger.New())

	api := r.App.Group("/api/v1")

	// public
	auth := api.Group("/auth")
	auth.Get("/:provider", r.AuthHandler.OAuthLogin)
	auth.Get("/:provider/callback", r.AuthHandler.OAuthCallback)
	auth.Post("/register", r.AuthHandler.Register)
	auth.Post("/login", r.AuthHandler.Login)
	api.Get("/plans", r.PlanHandler.GetAllPlans)

	// onboarding — partial token
	auth.Post("/onboarding", r.AuthMiddleware.OnboardingAuth, r.AuthHandler.CompleteOnboarding)

	// logout — full token
	auth.Post("/logout", r.AuthMiddleware.Auth, r.AuthHandler.Logout)

	// protected
	users := api.Group("/users", r.AuthMiddleware.Auth)
	users.Get("/me", r.UserHandler.GetMe)

	current := api.Group("/workspaces/current", r.AuthMiddleware.Auth)

	current.Get("/", r.WspHandler.GetCurrentWorkspace)
	current.Patch("/", r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.UpdateName)

	// templates
	current.Get("/templates", r.TemplateHandler.List)
	current.Post("/templates", r.TemplateHandler.Create)
	current.Get("/templates/:templateID", r.TemplateHandler.GetByID)
	current.Patch("/templates/:templateID", r.TemplateHandler.Update)
	current.Delete("/templates/:templateID", r.TemplateHandler.Delete)

	// template channels
	current.Post("/templates/:templateID/channels", r.TemplateHandler.CreateChannel)
	current.Get("/templates/:templateID/channels", r.TemplateHandler.ListChannels)
	current.Patch("/templates/:templateID/channels/:channelID", r.TemplateHandler.UpdateChannel)
	current.Delete("/templates/:templateID/channels/:channelID", r.TemplateHandler.DeleteChannel)

	// layouts
	current.Get("/layouts", r.LayoutHandler.List)
	current.Post("/layouts", r.LayoutHandler.Create)
	current.Get("/layouts/:id", r.LayoutHandler.GetByID)
	current.Patch("/layouts/:id", r.LayoutHandler.Update)
	current.Delete("/layouts/:id", r.LayoutHandler.Delete)
	current.Patch("/layouts/:id/default", r.LayoutHandler.SetDefault)
	
	// channel configs
	current.Get("/channels", r.ChnlConfigHandler.List)
	current.Post("/channels", r.ChnlConfigHandler.Create)
	current.Get("/channels/:id", r.ChnlConfigHandler.GetByID)
	current.Patch("/channels/:id", r.ChnlConfigHandler.Update)
	current.Delete("/channels/:id", r.ChnlConfigHandler.Delete)
	current.Patch("/channels/:id/default", r.ChnlConfigHandler.SetDefault)

	// members
	current.Get("/members", r.WspHandler.GetWorkspaceMembers)
	current.Patch("/members/:userID/role", r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.UpdateMemberRole)
	current.Delete("/members/:userID", r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.RemoveMember)
	
	// billing
	current.Get("/usage", r.BillingHandler.GetUsage)
	current.Get("/subscription", r.BillingHandler.GetSubscription)
	current.Delete("/subscription", r.AuthMiddleware.RequireRole("owner", "admin"), r.BillingHandler.CancelSubscription)
	current.Post("/checkout", r.AuthMiddleware.RequireRole("owner", "admin"), r.BillingHandler.CreateCheckout)




	apiKeys := current.Group("/api-keys")

	apiKeys.Get("/", r.ApiKeyHandler.ListAPIKeys)
	apiKeys.Post("/", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.CreateAPIKey)
	apiKeys.Delete("/:keyID", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.DeleteAPIKey)
	apiKeys.Patch("/:keyID/revoke", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.RevokeAPIKey)

	events := api.Group("/events", r.ApiKeyMiddleware.Authenticate)
	events.Post("/trigger", r.NotifHandler.Trigger)

	subscribers := api.Group("/identify", r.ApiKeyMiddleware.Authenticate)
	subscribers.Post("/", r.SubscriberHandler.Identify)
	subscribers.Post("/preferences", r.SubscriberHandler.UpsertPreference)
		// subscribers
	subscribers.Get("/subscribers", r.SubscriberHandler.List)
	subscribers.Delete("/subscribers/:id", r.SubscriberHandler.Delete)

}
