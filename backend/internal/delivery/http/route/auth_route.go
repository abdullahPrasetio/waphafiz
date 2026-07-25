package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterAuthRoutes(v1 fiber.Router, h *handler.AuthHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	authGroup := v1.Group("/auth")
	authGroup.Post("/register", h.Register)
	authGroup.Post("/register/family", h.RegisterFamily)
	authGroup.Post("/login", h.Login)
	authGroup.Post("/change-password", auth.Middleware(jwtCfg), userCtx, h.ChangePassword)
}
