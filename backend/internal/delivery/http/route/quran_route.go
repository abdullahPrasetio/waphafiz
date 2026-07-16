package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

func RegisterQuranRoutes(v1 fiber.Router, h *handler.QuranHandler, jwtCfg *auth.Config, userCtx fiber.Handler) {
	q := v1.Group("/quran", auth.Middleware(jwtCfg), userCtx)
	q.Get("/surah", h.GetSurahList)
	q.Get("/surah/:number", h.GetSurahDetail)
	q.Get("/ayah/:surah/:ayat/audio", h.GetAyahAudio)
}
