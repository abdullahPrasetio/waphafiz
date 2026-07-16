package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
)

func RegisterAuthRoutes(v1 fiber.Router, h *handler.AuthHandler) {
	auth := v1.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/register/family", h.RegisterFamily)
	auth.Post("/login", h.Login)
}
