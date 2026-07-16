package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
)

type MurajaahHandler struct {
	uc usecase.MurajaahUseCase
}

func NewMurajaahHandler(uc usecase.MurajaahUseCase) *MurajaahHandler {
	return &MurajaahHandler{uc: uc}
}

func (h *MurajaahHandler) GetToday(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	schedules, err := h.uc.GetToday(c.UserContext(), userID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "jadwal muraja'ah hari ini", schedules)
}

func (h *MurajaahHandler) MarkComplete(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id jadwal tidak valid")
	}

	if err := h.uc.MarkComplete(c.UserContext(), userID, scheduleID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			return response.NotFound(c, "jadwal tidak ditemukan")
		case errors.Is(err, usecase.ErrNotOwner):
			return response.Forbidden(c)
		}
		return response.InternalError(c)
	}
	return response.Success(c, "muraja'ah selesai", nil)
}

func (h *MurajaahHandler) GetHistory(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	schedules, total, err := h.uc.GetHistory(c.UserContext(), userID, page, limit)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Paginated(c, "riwayat muraja'ah", schedules, page, limit, int(total))
}
