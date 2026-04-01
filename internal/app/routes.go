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

	// onboarding — partial token
	auth.Post("/onboarding", r.AuthMiddleware.OnboardingAuth, r.AuthHandler.CompleteOnboarding)

	// logout — full token
	auth.Post("/logout", r.AuthMiddleware.Auth, r.AuthHandler.Logout)

	// protected
	users := api.Group("/users", r.AuthMiddleware.Auth)
	users.Get("/me", r.UserHandler.GetMe)

	workspaces := api.Group("/workspaces", r.AuthMiddleware.Auth)
	workspaces.Get("/current", r.WspHandler.GetCurrentWorkspace)
	workspaces.Patch("/current", r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.UpdateName)

	members := workspaces.Group("/current/members")
    members.Get("/", r.WspHandler.GetWorkspaceMembers)
	members.Patch("/:userID/role",r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.UpdateMemberRole)
	members.Delete("/:userID",r.AuthMiddleware.RequireRole("owner", "admin"), r.WspHandler.RemoveMember)


	apiKeys := api.Group("/workspaces/current/api-keys", r.AuthMiddleware.Auth)

	apiKeys.Get("/", r.ApiKeyHandler.ListAPIKeys)
	apiKeys.Post("/", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.CreateAPIKey)
	apiKeys.Delete("/:keyID", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.DeleteAPIKey)
	apiKeys.Patch("/:keyID/revoke", r.AuthMiddleware.RequireRole("owner", "admin"), r.ApiKeyHandler.RevokeAPIKey)

}
