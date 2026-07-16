package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
)

type DashboardHandler struct {
	uc usecase.DashboardUseCase
}

func NewDashboardHandler(uc usecase.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

func (h *DashboardHandler) GetMy(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	dash, err := h.uc.GetMyDashboard(c.UserContext(), userID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "dashboard", dash)
}

func (h *DashboardHandler) GetAdmin(c *fiber.Ctx) error {
	familyGroupID := mw.GetFamilyGroupID(c)
	if familyGroupID == uuid.Nil {
		return response.BadRequest(c, "user tidak tergabung dalam keluarga")
	}

	dash, err := h.uc.GetAdminDashboard(c.UserContext(), familyGroupID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "admin dashboard", dash)
}
