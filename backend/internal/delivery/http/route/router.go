package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
	"github.com/abdullahPrasetio/waphafiz/pkg/observability"
)

type Handlers struct {
	User      *handler.UserHandler
	Health    *handler.HealthHandler
	Auth      *handler.AuthHandler
	Family    *handler.FamilyHandler
	Quran     *handler.QuranHandler
	Hafalan   *handler.HafalanHandler
	Murajaah  *handler.MurajaahHandler
	Dashboard *handler.DashboardHandler
	Share     *handler.ShareHandler
}

func Setup(app *fiber.App, h *Handlers, jwtCfg *auth.Config, appEnv string, userCtx fiber.Handler) {
	app.Get("/", welcomeHandler(appEnv))
	app.Get("/health", h.Health.Check)
	app.Get("/metrics", prodGuard(appEnv), observability.MetricsHandler())

	v1 := app.Group("/api/v1")

	RegisterAuthRoutes(v1, h.Auth)
	RegisterUserRoutes(v1, h.User)
	RegisterQuranRoutes(v1, h.Quran, jwtCfg, userCtx)
	RegisterHafalanRoutes(v1, h.Hafalan, jwtCfg, userCtx)
	RegisterFamilyRoutes(v1, h.Family, jwtCfg, userCtx)
	RegisterMurajaahRoutes(v1, h.Murajaah, jwtCfg, userCtx)
	RegisterDashboardRoutes(v1, h.Dashboard, jwtCfg, userCtx)
	RegisterShareRoutes(v1, h.Share, jwtCfg, userCtx)
}

func prodGuard(appEnv string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if appEnv == "production" {
			return fiber.ErrNotFound
		}
		return c.Next()
	}
}

func welcomeHandler(appEnv string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		links := fiber.Map{"health": "/health"}
		if appEnv != "production" {
			links["metrics"] = "/metrics"
		}
		return c.JSON(fiber.Map{
			"service": "waphafiz",
			"version": "1.0.0",
			"env":     appEnv,
			"links":   links,
		})
	}
}
