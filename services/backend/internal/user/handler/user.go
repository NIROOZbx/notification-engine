package handler

import (
	userSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/response"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userSvc userSvc.UserService
}

func NewUserHandler(svc userSvc.UserService) *UserHandler {
    return &UserHandler{userSvc: svc}
}

func (u *UserHandler) CreateUser(ctx fiber.Ctx) error{
return response.OK(ctx,"","")

}