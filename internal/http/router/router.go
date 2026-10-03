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
	Config    *config.Config
	Logger    zerolog.Logger
	Redis     *redis.Client
	JWT       *jwtutil.Manager
	Auth      *handler.AuthHandler
	User      *handler.UserHandler
	Journal   *handler.JournalHandler
	Mood      *handler.MoodHandler
	Learning  *handler.LearningHandler
	SafeReach *handler.SafeReachHandler
	Community *handler.CommunityHandler
	Billing   *handler.BillingHandler
	Assistant *handler.AssistantHandler
	Health    *handler.HealthHandler
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
		middleware.SecurityHeaders(),
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
	registerLearningRoutes(api, deps)
	registerSafeReachRoutes(api, deps)
	registerCommunityRoutes(api, deps)
	registerBillingRoutes(api, deps)
	registerAssistantRoutes(api, deps)
	registerCertificateRoutes(api, deps)

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, deps Dependencies) {
	auth := api.Group("/auth")
	// A tighter, independent budget than general API traffic — these
	// endpoints are the ones worth brute-forcing (login, OTP, password
	// reset), so they get their own ceiling.
	auth.Use(middleware.RateLimit(deps.Redis, "auth", deps.Config.HTTP.AuthRateLimitPerMinute, time.Minute, deps.Logger))

	auth.POST("/register", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.Auth.Register)
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

	journal.POST("/entries", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.Journal.Create)
	journal.GET("/entries", deps.Journal.List)
	journal.GET("/entries/:id", deps.Journal.Get)
	journal.PATCH("/entries/:id", deps.Journal.Update)
	journal.DELETE("/entries/:id", deps.Journal.Delete)
	journal.GET("/export", deps.Journal.Export)
}

func registerMoodRoutes(api *gin.RouterGroup, deps Dependencies) {
	mood := api.Group("/mood")
	mood.Use(middleware.RequireAuth(deps.JWT))

	mood.POST("/entries", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.Mood.Create)
	mood.GET("/entries", deps.Mood.List)
	mood.GET("/entries/:id", deps.Mood.Get)
	mood.DELETE("/entries/:id", deps.Mood.Delete)
	mood.GET("/summary", deps.Mood.Summary)
}

func registerLearningRoutes(api *gin.RouterGroup, deps Dependencies) {
	learning := api.Group("/learning")
	learning.Use(middleware.RequireAuth(deps.JWT))

	learning.GET("/levels", deps.Learning.ListLevels)
	learning.GET("/levels/:id/lessons", deps.Learning.ListLessons)
	learning.GET("/lessons/:id", deps.Learning.GetLesson)
	learning.POST("/lessons/:id/complete", deps.Learning.CompleteLesson)
	learning.GET("/lessons/:id/quiz", deps.Learning.GetQuiz)
	learning.GET("/lessons/:id/audio", deps.Learning.GetLessonAudio)
	learning.POST("/quizzes/:id/submit", deps.Learning.SubmitQuiz)
	learning.GET("/progress", deps.Learning.GetProgress)
	learning.GET("/badges", deps.Learning.ListBadges)
	learning.GET("/levels/:id/certificate", deps.Learning.GetCertificate)
}

func registerCertificateRoutes(api *gin.RouterGroup, deps Dependencies) {
	certificates := api.Group("/certificates")

	// Public, unauthenticated — proving a certificate is genuine must not
	// require an account of your own.
	certificates.GET("/verify/:serial", deps.Learning.VerifyCertificate)
}

func registerSafeReachRoutes(api *gin.RouterGroup, deps Dependencies) {
	safereach := api.Group("/safereach")

	// Public, token-gated — registered before RequireAuth is attached to
	// the group below. A gin group's Use() only affects routes
	// registered after the call, so this one route alone skips it.
	safereach.GET("/public/alerts/:id", deps.SafeReach.PublicView)

	safereach.Use(middleware.RequireAuth(deps.JWT))

	safereach.POST("/contacts", deps.SafeReach.RegisterContact)
	safereach.GET("/contacts", deps.SafeReach.ListContacts)
	safereach.DELETE("/contacts/:id", deps.SafeReach.DeleteContact)

	safereach.POST("/alerts", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.SafeReach.TriggerAlert)
	safereach.GET("/alerts", deps.SafeReach.ListAlerts)
	safereach.GET("/alerts/:id", deps.SafeReach.GetAlert)
	safereach.POST("/alerts/:id/cancel", deps.SafeReach.CancelAlert)
	safereach.POST("/alerts/:id/resolve", deps.SafeReach.ResolveAlert)
	safereach.POST("/alerts/:id/location", deps.SafeReach.RecordLocation)
	safereach.GET("/alerts/:id/stream", deps.SafeReach.Stream)

	safereach.POST("/device-tokens", deps.SafeReach.RegisterDeviceToken)
}

func registerCommunityRoutes(api *gin.RouterGroup, deps Dependencies) {
	community := api.Group("/community")
	community.Use(middleware.RequireAuth(deps.JWT))

	community.POST("/posts", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.Community.CreatePost)
	community.GET("/posts", deps.Community.ListFeed)
	community.DELETE("/posts/:id", deps.Community.DeletePost)
	community.POST("/posts/:id/comments", middleware.Idempotency(deps.Redis, deps.Config.HTTP.IdempotencyTTL, deps.Logger), deps.Community.CreateComment)
	community.GET("/posts/:id/comments", deps.Community.ListComments)
	community.DELETE("/comments/:id", deps.Community.DeleteComment)
	community.POST("/reports", deps.Community.FileReport)

	admin := api.Group("/admin/community")
	admin.Use(middleware.RequireAuth(deps.JWT), middleware.RequireRole("admin"))
	admin.GET("/reports", deps.Community.ListOpenReports)
	admin.POST("/reports/:id/resolve", deps.Community.ResolveReport)
}

func registerBillingRoutes(api *gin.RouterGroup, deps Dependencies) {
	billing := api.Group("/billing")

	// Unauthenticated — Paystack calls this directly. Authenticity
	// comes entirely from the signature BillingHandler.Webhook verifies,
	// not from any session; registered before RequireAuth is attached
	// to the group below, so this one route alone skips it.
	billing.POST("/webhook", deps.Billing.Webhook)

	billing.Use(middleware.RequireAuth(deps.JWT))
	billing.POST("/subscribe", deps.Billing.Subscribe)
	billing.GET("/subscription", deps.Billing.GetSubscription)
}

func registerAssistantRoutes(api *gin.RouterGroup, deps Dependencies) {
	assistant := api.Group("/assistant")
	assistant.Use(middleware.RequireAuth(deps.JWT))

	assistant.POST("/messages", deps.Assistant.SendMessage)
	assistant.GET("/conversations", deps.Assistant.ListConversations)
	assistant.GET("/conversations/:id/messages", deps.Assistant.ListMessages)
}
