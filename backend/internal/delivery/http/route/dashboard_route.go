package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterDashboardRoutes(v1 fiber.Router, h *handler.DashboardHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	d := v1.Group("/dashboard", auth.Middleware(jwtCfg), userCtx)
	d.Get("/me", h.GetMy)

	admin := v1.Group("/admin", auth.Middleware(jwtCfg), userCtx, auth.RequireRole("admin"))
	admin.Get("/dashboard", h.GetAdmin)
}
