package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
	"github.com/abdullahPrasetio/waphafiz/pkg/validator"
)

type AuthHandler struct {
	uc  usecase.AuthUseCase
	val *validator.Validator
}

func NewAuthHandler(uc usecase.AuthUseCase, val *validator.Validator) *AuthHandler {
	return &AuthHandler{uc: uc, val: val}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req usecase.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	res, err := h.uc.Register(c.UserContext(), &req)
	if err != nil {
		return h.mapError(c, err)
	}
	return response.Created(c, "registrasi berhasil", res)
}

func (h *AuthHandler) RegisterFamily(c *fiber.Ctx) error {
	var req usecase.RegisterFamilyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	res, err := h.uc.RegisterFamily(c.UserContext(), &req)
	if err != nil {
		return h.mapError(c, err)
	}
	return response.Created(c, "keluarga berhasil dibuat", res)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req usecase.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}
	if err := h.val.Validate(&req); err != nil {
		return response.ValidationError(c, err.Error())
	}

	res, err := h.uc.Login(c.UserContext(), &req)
	if err != nil {
		return h.mapError(c, err)
	}
	return response.Success(c, "login berhasil", res)
}

func (h *AuthHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidInviteCode):
		return response.BadRequest(c, err.Error())
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return response.Error(c, fiber.StatusUnauthorized, response.ErrUnauthorized, err.Error())
	case errors.Is(err, usecase.ErrUserInactive):
		return response.Error(c, fiber.StatusForbidden, response.ErrForbidden, err.Error())
	case errors.Is(err, usecase.ErrEmailConflict):
		return response.Conflict(c, "email sudah terdaftar")
	default:
		log.Error().Err(err).Str("request_id", c.GetRespHeader("X-Request-Id")).Msg("auth handler: unhandled error")
		return response.InternalError(c)
	}
}
