package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterFamilyRoutes(v1 fiber.Router, h *handler.FamilyHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	family := v1.Group("/family", auth.Middleware(jwtCfg), userCtx, auth.RequireRole("admin"))
	family.Get("/members", h.ListMembers)
	family.Patch("/members/:id/status", h.UpdateMemberStatus)
	family.Post("/invite-code/regenerate", h.RegenerateInviteCode)
}
