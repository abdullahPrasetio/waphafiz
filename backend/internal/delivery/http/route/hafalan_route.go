package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterHafalanRoutes(v1 fiber.Router, h *handler.HafalanHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	protected := v1.Group("", auth.Middleware(jwtCfg), userCtx)

	hafalan := protected.Group("/hafalan")
	hafalan.Post("/", h.Create)
	hafalan.Get("/", h.List)
	hafalan.Get("/summary", h.Summary)
	hafalan.Patch("/:id", h.Update)
	hafalan.Delete("/:id", h.Delete)

	admin := protected.Group("/admin", auth.RequireRole("admin"))
	admin.Get("/hafalan", h.AdminList)
	admin.Get("/hafalan/member/:userID", h.AdminListByMember)
	admin.Patch("/hafalan/:id", h.AdminUpdate)
	admin.Delete("/hafalan/:id", h.AdminDelete)
}
