package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterHafalanRoutes(v1 fiber.Router, h *handler.HafalanHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	hafalan := v1.Group("/hafalan", auth.Middleware(jwtCfg), userCtx)
	hafalan.Post("/", h.Create)
	hafalan.Get("/", h.List)
	hafalan.Get("/summary", h.Summary)
	hafalan.Patch("/:id", h.Update)
	hafalan.Delete("/:id", h.Delete)

	admin := v1.Group("/admin", auth.Middleware(jwtCfg), userCtx, auth.RequireRole("admin"))
	admin.Get("/hafalan", h.AdminList)
	admin.Get("/hafalan/member/:userID", h.AdminListByMember)
	admin.Patch("/hafalan/:id", h.AdminUpdate)
	admin.Delete("/hafalan/:id", h.AdminDelete)
}
