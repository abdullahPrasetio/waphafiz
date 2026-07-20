package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
	"github.com/abdullahPrasetio/waphafiz/pkg/validator"
)

type ShareHandler struct {
	uc  usecase.ShareUseCase
	val *validator.Validator
}

func NewShareHandler(uc usecase.ShareUseCase, val *validator.Validator) *ShareHandler {
	return &ShareHandler{uc: uc, val: val}
}

func (h *ShareHandler) Create(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	var req usecase.CreateShareInput
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	res, err := h.uc.CreateShare(c.UserContext(), userID, mw.GetUserRole(c), mw.GetFamilyGroupID(c), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrShareForbidden):
			return response.Forbidden(c)
		case errors.Is(err, usecase.ErrShareLimitReached):
			return response.BadRequest(c, err.Error())
		case errors.Is(err, usecase.ErrShareUserNotFound):
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c)
	}
	return response.Created(c, "link pantau berhasil dibuat — token hanya ditampilkan sekali", res)
}

func (h *ShareHandler) List(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	var targetID *uuid.UUID
	if q := c.Query("user_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			return response.BadRequest(c, "user_id tidak valid")
		}
		targetID = &id
	}

	list, err := h.uc.ListShares(c.UserContext(), userID, mw.GetUserRole(c), mw.GetFamilyGroupID(c), targetID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrShareForbidden):
			return response.Forbidden(c)
		case errors.Is(err, usecase.ErrShareUserNotFound):
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c)
	}
	return response.Success(c, "daftar link pantau", list)
}

func (h *ShareHandler) Revoke(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	shareID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id tidak valid")
	}

	if err := h.uc.RevokeShare(c.UserContext(), userID, mw.GetUserRole(c), mw.GetFamilyGroupID(c), shareID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrShareForbidden):
			return response.Forbidden(c)
		case errors.Is(err, usecase.ErrShareNotFound), errors.Is(err, usecase.ErrShareUserNotFound):
			return response.NotFound(c, "link pantau tidak ditemukan")
		}
		return response.InternalError(c)
	}
	return response.Success(c, "link pantau dicabut", nil)
}

// ViewPublic adalah satu-satunya endpoint tanpa JWT selain /health.
// Semua kegagalan dijawab 404 identik — jangan bedakan token salah /
// dicabut / kedaluwarsa (lihat docs/design-share-dashboard.md §2.2).
func (h *ShareHandler) ViewPublic(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return response.NotFound(c, usecase.ErrShareNotFound.Error())
	}

	dash, err := h.uc.ViewShared(c.UserContext(), token)
	if err != nil {
		if errors.Is(err, usecase.ErrShareNotFound) {
			return response.NotFound(c, usecase.ErrShareNotFound.Error())
		}
		return response.InternalError(c)
	}
	return response.Success(c, "dashboard", dash)
}
