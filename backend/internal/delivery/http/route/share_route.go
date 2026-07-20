package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterShareRoutes(v1 fiber.Router, h *handler.ShareHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	s := v1.Group("/shares", auth.Middleware(jwtCfg), userCtx)
	s.Post("/", h.Create)
	s.Get("/", h.List)
	s.Delete("/:id", h.Revoke)

	// Endpoint publik tanpa JWT dengan rate limit lebih ketat.
	pub := v1.Group("/public", mw.PublicShareRateLimiter())
	pub.Get("/shares/:token", h.ViewPublic)
}
