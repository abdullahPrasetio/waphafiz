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

type HafalanHandler struct {
	uc  usecase.HafalanUseCase
	val *validator.Validator
}

func NewHafalanHandler(uc usecase.HafalanUseCase, val *validator.Validator) *HafalanHandler {
	return &HafalanHandler{uc: uc, val: val}
}

func (h *HafalanHandler) Create(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	var req usecase.CreateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	hafalan, err := h.uc.Create(c.UserContext(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrAyatInvalid):
			return response.BadRequest(c, err.Error())
		case errors.Is(err, usecase.ErrAyatOverlap):
			return response.Conflict(c, err.Error())
		}
		return response.InternalError(c)
	}
	return response.Created(c, "hafalan berhasil disimpan", hafalan)
}

func (h *HafalanHandler) List(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	list, err := h.uc.List(c.UserContext(), userID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "daftar hafalan", list)
}

func (h *HafalanHandler) Update(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	hafalanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id hafalan tidak valid")
	}

	var req usecase.UpdateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	hafalan, err := h.uc.Update(c.UserContext(), userID, hafalanID, &req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			return response.NotFound(c, "hafalan tidak ditemukan")
		case errors.Is(err, usecase.ErrNotOwner):
			return response.Forbidden(c)
		}
		return response.InternalError(c)
	}
	return response.Success(c, "hafalan diperbarui", hafalan)
}

func (h *HafalanHandler) Delete(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	hafalanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id hafalan tidak valid")
	}

	if err := h.uc.Delete(c.UserContext(), userID, hafalanID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			return response.NotFound(c, "hafalan tidak ditemukan")
		case errors.Is(err, usecase.ErrNotOwner):
			return response.Forbidden(c)
		}
		return response.InternalError(c)
	}
	return response.Success(c, "hafalan dihapus", nil)
}

func (h *HafalanHandler) AdminListByMember(c *fiber.Ctx) error {
	memberID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return response.BadRequest(c, "user id tidak valid")
	}

	list, err := h.uc.AdminListByMember(c.UserContext(), memberID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "daftar hafalan anggota", list)
}

func (h *HafalanHandler) AdminUpdate(c *fiber.Ctx) error {
	hafalanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id hafalan tidak valid")
	}

	var req usecase.UpdateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	hafalan, err := h.uc.AdminUpdate(c.UserContext(), hafalanID, &req)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			return response.NotFound(c, "hafalan tidak ditemukan")
		}
		return response.InternalError(c)
	}
	return response.Success(c, "hafalan diperbarui", hafalan)
}

func (h *HafalanHandler) AdminDelete(c *fiber.Ctx) error {
	hafalanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id hafalan tidak valid")
	}

	if err := h.uc.AdminDelete(c.UserContext(), hafalanID); err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			return response.NotFound(c, "hafalan tidak ditemukan")
		}
		return response.InternalError(c)
	}
	return response.Success(c, "hafalan dihapus", nil)
}

func (h *HafalanHandler) Summary(c *fiber.Ctx) error {
	userID := mw.GetUserID(c)
	if userID == uuid.Nil {
		return response.Unauthorized(c)
	}

	summary, err := h.uc.Summary(c.UserContext(), userID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "ringkasan hafalan", summary)
}

func (h *HafalanHandler) AdminList(c *fiber.Ctx) error {
	familyGroupID := mw.GetFamilyGroupID(c)
	if familyGroupID == uuid.Nil {
		return response.BadRequest(c, "user tidak tergabung dalam keluarga")
	}

	list, err := h.uc.ListByFamilyGroup(c.UserContext(), familyGroupID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "progress hafalan anggota", list)
}
