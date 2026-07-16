package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/response"
)

type FamilyHandler struct {
	uc         usecase.FamilyUseCase
	familyRepo domainrepo.FamilyGroupRepository
}

func NewFamilyHandler(uc usecase.FamilyUseCase, familyRepo domainrepo.FamilyGroupRepository) *FamilyHandler {
	return &FamilyHandler{uc: uc, familyRepo: familyRepo}
}

func (h *FamilyHandler) ListMembers(c *fiber.Ctx) error {
	familyGroupID := mw.GetFamilyGroupID(c)
	if familyGroupID == uuid.Nil {
		return response.BadRequest(c, "user tidak tergabung dalam keluarga")
	}

	fg, err := h.familyRepo.FindByID(c.UserContext(), familyGroupID)
	if err != nil {
		return response.InternalError(c)
	}

	members, err := h.uc.ListMembers(c.UserContext(), familyGroupID)
	if err != nil {
		return response.InternalError(c)
	}

	return response.Success(c, "anggota berhasil dimuat", fiber.Map{
		"members":     members,
		"invite_code": fg.InviteCode,
	})
}

func (h *FamilyHandler) UpdateMemberStatus(c *fiber.Ctx) error {
	familyGroupID := mw.GetFamilyGroupID(c)
	if familyGroupID == uuid.Nil {
		return response.BadRequest(c, "user tidak tergabung dalam keluarga")
	}

	memberID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "id member tidak valid")
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "request tidak valid")
	}

	if err := h.uc.UpdateMemberStatus(c.UserContext(), familyGroupID, memberID, body.IsActive); err != nil {
		if errors.Is(err, usecase.ErrMemberNotInFamily) {
			return response.NotFound(c, "member tidak ditemukan")
		}
		return response.InternalError(c)
	}
	return response.Success(c, "status anggota diperbarui", nil)
}

func (h *FamilyHandler) RegenerateInviteCode(c *fiber.Ctx) error {
	familyGroupID := mw.GetFamilyGroupID(c)
	if familyGroupID == uuid.Nil {
		return response.BadRequest(c, "user tidak tergabung dalam keluarga")
	}

	code, err := h.uc.RegenerateInviteCode(c.UserContext(), familyGroupID)
	if err != nil {
		return response.InternalError(c)
	}
	return response.Success(c, "kode undangan diperbarui", fiber.Map{"invite_code": code})
}
