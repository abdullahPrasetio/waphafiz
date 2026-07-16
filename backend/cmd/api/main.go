package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/abdullahPrasetio/waphafiz/config"
	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/handler"
	mw "github.com/abdullahPrasetio/waphafiz/internal/delivery/http/middleware"
	"github.com/abdullahPrasetio/waphafiz/internal/delivery/http/route"
	"github.com/abdullahPrasetio/waphafiz/internal/domain/entity"
	dbrepo "github.com/abdullahPrasetio/waphafiz/internal/repository/db"
	"github.com/abdullahPrasetio/waphafiz/internal/usecase"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
	"github.com/abdullahPrasetio/waphafiz/pkg/database"
	applogger "github.com/abdullahPrasetio/waphafiz/pkg/logger"
	"github.com/abdullahPrasetio/waphafiz/pkg/observability"
	"github.com/abdullahPrasetio/waphafiz/pkg/quranapi"
	"github.com/abdullahPrasetio/waphafiz/pkg/validator"
)

const version = "0.1.0"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	applogger.Setup(cfg.App.Env, cfg.Log.Level, cfg.Log.FilePath, cfg.Log.ToFile, cfg.App.Name)

	if err := applogger.SetupSinks(applogger.SinkConfig{
		Dir:        cfg.Log.Dir,
		Rotation:   cfg.Log.Rotation,
		MaxAgeDays: cfg.Log.MaxAgeDays,
		Console:    cfg.App.Env != "production",
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to setup log sinks")
	}

	log.Info().Str("version", version).Str("env", cfg.App.Env).Msg("starting waphafiz")

	obsProvider, err := observability.New(context.Background(), &cfg.Observability, cfg.App.Name, version, cfg.App.Env)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup observability provider")
	}

	db, err := database.NewConnection(&cfg.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	if err := obsProvider.InstrumentGORM(db); err != nil {
		log.Warn().Err(err).Msg("gorm instrumentation failed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get underlying sql.DB")
	}

	if cfg.DB.AutoMigrate {
		if err := db.AutoMigrate(
			&entity.FamilyGroup{},
			&entity.User{},
			&entity.HafalanProgress{},
			&entity.MurajaahSchedule{},
			&entity.MurajaahLog{},
		); err != nil {
			log.Fatal().Err(err).Msg("auto-migrate failed")
		}
	}

	redisClient := newRedisClient(&cfg.Redis)
	obsProvider.InstrumentRedis(redisClient)

	// JWT config
	jwtCfg := &auth.Config{
		Secret:   cfg.JWT.Secret,
		Issuer:   cfg.JWT.Issuer,
		Audience: cfg.JWT.Audience,
	}
	if d, err := time.ParseDuration(cfg.JWT.Expiry); err == nil {
		jwtCfg.Expiry = d
	}

	// Repositories
	userRepo := dbrepo.NewUserRepository(db)
	familyRepo := dbrepo.NewFamilyGroupRepository(db)
	hafalanRepo := dbrepo.NewHafalanRepository(db)
	murajaahRepo := dbrepo.NewMurajaahRepository(db)

	// Usecases
	authUC := usecase.NewAuthUseCase(userRepo, familyRepo, jwtCfg)
	familyUC := usecase.NewFamilyUseCase(familyRepo)
	quranClient := quranapi.New(redisClient)
	quranUC := usecase.NewQuranUseCase(quranClient)
	hafalanUC := usecase.NewHafalanUseCase(hafalanRepo)
	murajaahUC := usecase.NewMurajaahUseCase(murajaahRepo, hafalanRepo)
	dashboardUC := usecase.NewDashboardUseCase(hafalanRepo, familyRepo, murajaahRepo)
	userUC := usecase.NewUserUseCase(userRepo)

	val := validator.New()
	startTime := time.Now()

	// Handlers
	handlers := &route.Handlers{
		User:      handler.NewUserHandler(userUC, val),
		Health:    handler.NewHealthHandler(sqlDB, startTime, version),
		Auth:      handler.NewAuthHandler(authUC, val),
		Family:    handler.NewFamilyHandler(familyUC, familyRepo),
		Quran:     handler.NewQuranHandler(quranUC),
		Hafalan:   handler.NewHafalanHandler(hafalanUC, val),
		Murajaah:  handler.NewMurajaahHandler(murajaahUC),
		Dashboard: handler.NewDashboardHandler(dashboardUC),
	}

	handlers.Health.AddChecker("redis", func(ctx context.Context) string {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			return "down"
		}
		return "ok"
	})

	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		BodyLimit:             4 * 1024 * 1024,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           120 * time.Second,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  false,
				"message": http.StatusText(code),
			})
		},
	})

	app.Use(mw.Recover())
	app.Use(mw.RequestID())
	app.Use(mw.SecurityHeaders())
	app.Use(mw.RateLimiter())
	app.Use(mw.RequestLogger())
	app.Use(mw.CORS(cfg.App.CORSAllowedOrigins))
	app.Use(obsProvider.HTTPMiddleware())
	app.Use(observability.MetricsMiddleware())
	app.Use(mw.AccessLog(mw.AccessLogOptions{
		BodyMaxBytes:  cfg.Log.BodyMaxBytes,
		CaptureBodies: cfg.Log.HTTPBodies,
	}))
	// Applied per-route after auth.Middleware so JWT claims are already verified.
	userCtx := mw.UserContext(userRepo)

	route.Setup(app, handlers, jwtCfg, cfg.App.Env, userCtx)

	go func() {
		addr := ":" + cfg.App.Port
		log.Info().Str("addr", addr).Msg("http server listening")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("http server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutdown signal received")
	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutCtx); err != nil {
		log.Error().Err(err).Msg("http server forced shutdown")
	}
	if err := obsProvider.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("observability provider shutdown error")
	}

	redisClient.Close() //nolint:errcheck
	sqlDB.Close()       //nolint:errcheck

	log.Info().Msg("shutdown complete")
}

func newRedisClient(cfg *config.RedisConfig) *redis.Client {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		opts = &redis.Options{Addr: "localhost:6379"}
	}
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB > 0 {
		opts.DB = cfg.DB
	}
	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}
	if cfg.MaxRetries > 0 {
		opts.MaxRetries = cfg.MaxRetries
	}
	if d, err := time.ParseDuration(cfg.DialTimeout); err == nil {
		opts.DialTimeout = d
	}
	if d, err := time.ParseDuration(cfg.ReadTimeout); err == nil {
		opts.ReadTimeout = d
	}
	if d, err := time.ParseDuration(cfg.WriteTimeout); err == nil {
		opts.WriteTimeout = d
	}
	return redis.NewClient(opts)
}
