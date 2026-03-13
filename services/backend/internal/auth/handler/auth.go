package handler

import (
	"github.com/NIROOZbx/notification-engine/services/backend/internal/auth/services"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	"github.com/NIROOZbx/notification-engine/services/pkg/jwt"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/shareed2k/goth_fiber/v2"
)



type AuthHandler struct {
	service services.AuthService
}

func (h *AuthHandler) OAuthLogin(c fiber.Ctx) error {
	return goth_fiber.BeginAuthHandler(c)

}
func (h *AuthHandler) OAuthCallback(c fiber.Ctx) error {

	gothUser, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication failed")
	}
	userDetails := &dtos.OAuthUserDetails{
		Name:  gothUser.Name,
		Email: gothUser.Email,
		AvatarURL: gothUser.AvatarURL,
		Provider: gothUser.Provider,
		UserID:   gothUser.UserID,
	
	}

	user,tokenPair,err:=h.service.HandleOAuthCallback(c.Context(), userDetails)
	if err!=nil{
		return response.BadRequest(c,nil,"Failed to process user data")
	}



	h.setTokenCookies(c,tokenPair)


	return response.OK(c,"sucessful login",user)

	
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}


func (h *AuthHandler)setTokenCookies(c fiber.Ctx,tokenPair *jwt.Pair){

}