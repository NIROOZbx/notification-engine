package app

import ()

func SetUpRoutes(r *RouterDeps) {
	api := r.App.Group("/api/v1")

	// public
	auth := api.Group("/auth")
	auth.Get("/google", r.AuthHandler.OAuthLogin)
	auth.Get("/google/callback", r.AuthHandler.OAuthCallback)

	// onboarding — partial token
	auth.Post("/onboarding", r.AuthMiddleware.OnboardingAuth, r.AuthHandler.CompleteOnboarding)

	// logout — full token
	auth.Post("/logout", r.AuthMiddleware.Auth, r.AuthHandler.Logout)

	// protected
	users := api.Group("/users", r.AuthMiddleware.Auth)
	users.Get("/me", r.UserHandler.GetMe)

	workspaces := api.Group("/workspaces", r.AuthMiddleware.Auth)
	workspaces.Get("/current", r.WspHandler.GetCurrentWorkspace)
}
