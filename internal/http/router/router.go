// Package router assembles the Gin engine: global middleware, route
// groups, and their per-group middleware (auth, rate limiting). It is
// the one place that imports both internal/http/handler and
// internal/http/middleware, so it necessarily lives outside package
// http itself (handler and middleware both import that package for the
// response envelope, and Go disallows the resulting import cycle).
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/handler"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
)

// Dependencies bundles everything the router needs to wire routes to
// handlers and attach the right middleware to each group.
type Dependencies struct {
	Config  *config.Config
	Logger  zerolog.Logger
	Redis   *redis.Client
	JWT     *jwtutil.Manager
	Auth    *handler.AuthHandler
	User    *handler.UserHandler
	Journal *handler.JournalHandler
	Mood    *handler.MoodHandler
	Health  *handler.HealthHandler
}

func New(deps Dependencies) *gin.Engine {
	if deps.Config.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logging(deps.Logger),
		middleware.CORS(deps.Config.HTTP.AllowedOrigins),
	)

	r.GET("/healthz", deps.Health.Liveness)
	r.GET("/readyz", deps.Health.Readiness)

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit(deps.Redis, "general", deps.Config.HTTP.RateLimitPerMinute, time.Minute, deps.Logger))

	registerAuthRoutes(api, deps)
	registerUserRoutes(api, deps)
	registerJournalRoutes(api, deps)
	registerMoodRoutes(api, deps)

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, deps Dependencies) {
	auth := api.Group("/auth")
	// A tighter, independent budget than general API traffic — these
	// endpoints are the ones worth brute-forcing (login, OTP, password
	// reset), so they get their own ceiling.
	auth.Use(middleware.RateLimit(deps.Redis, "auth", deps.Config.HTTP.AuthRateLimitPerMinute, time.Minute, deps.Logger))

	auth.POST("/register", deps.Auth.Register)
	auth.POST("/verify-email", deps.Auth.VerifyEmail)
	auth.POST("/verify-phone", deps.Auth.VerifyPhone)
	auth.POST("/resend-otp", deps.Auth.ResendOTP)
	auth.POST("/login", deps.Auth.Login)
	auth.POST("/refresh", deps.Auth.Refresh)
	auth.POST("/logout", deps.Auth.Logout)
	auth.POST("/forgot-password", deps.Auth.ForgotPassword)
	auth.POST("/reset-password", deps.Auth.ResetPassword)
}

func registerUserRoutes(api *gin.RouterGroup, deps Dependencies) {
	users := api.Group("/users")
	users.Use(middleware.RequireAuth(deps.JWT))

	users.GET("/me", deps.User.Me)
	users.PATCH("/me", deps.User.UpdateMe)
}

func registerJournalRoutes(api *gin.RouterGroup, deps Dependencies) {
	journal := api.Group("/journal")
	journal.Use(middleware.RequireAuth(deps.JWT))

	journal.POST("/entries", deps.Journal.Create)
	journal.GET("/entries", deps.Journal.List)
	journal.GET("/entries/:id", deps.Journal.Get)
	journal.PATCH("/entries/:id", deps.Journal.Update)
	journal.DELETE("/entries/:id", deps.Journal.Delete)
	journal.GET("/export", deps.Journal.Export)
}

func registerMoodRoutes(api *gin.RouterGroup, deps Dependencies) {
	mood := api.Group("/mood")
	mood.Use(middleware.RequireAuth(deps.JWT))

	mood.POST("/entries", deps.Mood.Create)
	mood.GET("/entries", deps.Mood.List)
	mood.GET("/entries/:id", deps.Mood.Get)
	mood.DELETE("/entries/:id", deps.Mood.Delete)
	mood.GET("/summary", deps.Mood.Summary)
}
