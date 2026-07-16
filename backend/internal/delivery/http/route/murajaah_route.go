package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterMurajaahRoutes(v1 fiber.Router, h *handler.MurajaahHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	m := v1.Group("/murajaah", auth.Middleware(jwtCfg), userCtx)
	m.Get("/today", h.GetToday)
	m.Get("/history", h.GetHistory)
	m.Post("/:id/complete", h.MarkComplete)
}
