package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
)

type QuranHandler struct {
	uc usecase.QuranUseCase
}

func NewQuranHandler(uc usecase.QuranUseCase) *QuranHandler {
	return &QuranHandler{uc: uc}
}

func (h *QuranHandler) GetSurahList(c *fiber.Ctx) error {
	surahs, err := h.uc.GetSurahList(c.UserContext())
	if err != nil {
		return response.Error(c, fiber.StatusServiceUnavailable, response.ErrInternal, "Quran API tidak tersedia")
	}
	return response.Success(c, "surah list", surahs)
}

func (h *QuranHandler) GetSurahDetail(c *fiber.Ctx) error {
	number, err := strconv.Atoi(c.Params("number"))
	if err != nil {
		return response.BadRequest(c, "nomor surah tidak valid")
	}

	detail, err := h.uc.GetSurahDetail(c.UserContext(), number)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			return response.NotFound(c, "surah tidak ditemukan")
		}
		return response.Error(c, fiber.StatusServiceUnavailable, response.ErrInternal, "Quran API tidak tersedia")
	}
	return response.Success(c, "surah detail", detail)
}

func (h *QuranHandler) GetAyahAudio(c *fiber.Ctx) error {
	surah, err := strconv.Atoi(c.Params("surah"))
	if err != nil {
		return response.BadRequest(c, "nomor surah tidak valid")
	}
	ayah, err := strconv.Atoi(c.Params("ayat"))
	if err != nil {
		return response.BadRequest(c, "nomor ayat tidak valid")
	}

	audio, err := h.uc.GetAyahAudio(c.UserContext(), surah, ayah)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) {
			return response.NotFound(c, "ayat tidak ditemukan")
		}
		return response.Error(c, fiber.StatusServiceUnavailable, response.ErrInternal, "Quran API tidak tersedia")
	}
	return response.Success(c, "ayah audio", audio)
}
