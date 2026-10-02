// Command api runs the TraumaCompendium HTTP API.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/config"
	"github.com/traumacompendium/backend/internal/http/handler"
	"github.com/traumacompendium/backend/internal/http/router"
	"github.com/traumacompendium/backend/internal/notification/ai"
	"github.com/traumacompendium/backend/internal/notification/email"
	"github.com/traumacompendium/backend/internal/notification/paystack"
	"github.com/traumacompendium/backend/internal/notification/push"
	"github.com/traumacompendium/backend/internal/notification/sms"
	"github.com/traumacompendium/backend/internal/notification/tts"
	"github.com/traumacompendium/backend/internal/pkg/jwtutil"
	"github.com/traumacompendium/backend/internal/platform/cache"
	"github.com/traumacompendium/backend/internal/platform/database"
	"github.com/traumacompendium/backend/internal/platform/logger"
	"github.com/traumacompendium/backend/internal/repository/mysql"
	"github.com/traumacompendium/backend/internal/service/assistantservice"
	"github.com/traumacompendium/backend/internal/service/authservice"
	"github.com/traumacompendium/backend/internal/service/billingservice"
	"github.com/traumacompendium/backend/internal/service/communityservice"
	"github.com/traumacompendium/backend/internal/service/journalservice"
	"github.com/traumacompendium/backend/internal/service/learningservice"
	"github.com/traumacompendium/backend/internal/service/moodservice"
	"github.com/traumacompendium/backend/internal/service/safereachservice"
	"github.com/traumacompendium/backend/internal/service/userservice"
	"github.com/traumacompendium/backend/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App)
	log.Info().Str("env", cfg.App.Env).Msg("starting traumacompendium api")

	// startupCtx bounds dependency connection/migration at boot; the
	// server's own request handling uses a separate, long-lived context
	// tied to OS signals (below).
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelStartup()

	db, err := database.Open(startupCtx, cfg.Database, log)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close() //nolint:errcheck // best-effort on process shutdown

	var migrationsFS fs.FS = migrations.FS
	if err := database.Migrate(startupCtx, db, migrationsFS, log); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	redisClient, err := cache.Open(startupCtx, cfg.Redis, log)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer redisClient.Close() //nolint:errcheck // best-effort on process shutdown

	srv := buildServer(cfg, log, db, redisClient)

	return serveWithGracefulShutdown(srv, cfg.HTTP.ShutdownGracePeriod, log)
}

func buildServer(cfg *config.Config, log zerolog.Logger, db *sql.DB, redisClient *redis.Client) *nethttp.Server {
	userRepo := mysql.NewUserRepository(db)
	otpRepo := mysql.NewOTPRepository(db)
	refreshTokenRepo := mysql.NewRefreshTokenRepository(db)
	journalRepo := mysql.NewJournalRepository(db)
	moodRepo := mysql.NewMoodRepository(db)
	learningContentRepo := mysql.NewLearningContentRepository(db)
	learningProgressRepo := mysql.NewLearningProgressRepository(db)
	learningAudioRepo := mysql.NewLearningAudioRepository(db)
	safeReachContactRepo := mysql.NewSafeReachContactRepository(db)
	safeReachAlertRepo := mysql.NewSafeReachAlertRepository(db)
	deviceTokenRepo := mysql.NewDeviceTokenRepository(db)
	communityPostRepo := mysql.NewCommunityPostRepository(db)
	communityCommentRepo := mysql.NewCommunityCommentRepository(db)
	communityReportRepo := mysql.NewCommunityReportRepository(db)
	subscriptionRepo := mysql.NewSubscriptionRepository(db)
	webhookEventRepo := mysql.NewWebhookEventRepository(db)
	assistantRepo := mysql.NewAssistantRepository(db)

	jwtManager := jwtutil.NewManager(cfg.JWT)
	emailSender := email.New(cfg.SMTP, log)
	smsSender := sms.New(cfg.Termii, log)
	ttsSynth := tts.New(cfg.TTS, log)
	pushSender := push.New(cfg.Push, log)
	publisher := cache.Publisher{Client: redisClient}
	aiClient := ai.New(cfg.AI, log)
	paystackClient := paystack.New(cfg.Paystack, log)

	authSvc := authservice.New(userRepo, otpRepo, refreshTokenRepo, jwtManager, emailSender, smsSender, cfg.OTP, log)
	userSvc := userservice.New(userRepo)
	journalSvc := journalservice.New(journalRepo)
	moodSvc := moodservice.New(moodRepo)
	learningSvc := learningservice.New(learningContentRepo, learningProgressRepo, learningAudioRepo, ttsSynth)
	safeReachSvc := safereachservice.New(
		safeReachContactRepo, safeReachAlertRepo, deviceTokenRepo,
		smsSender, emailSender, pushSender, publisher,
		cfg.App.BaseURL, log,
	)
	communitySvc := communityservice.New(communityPostRepo, communityCommentRepo, communityReportRepo, aiClient, log)
	billingSvc := billingservice.New(subscriptionRepo, webhookEventRepo, userRepo, paystackClient, cfg.Paystack.SecretKey, cfg.App.BaseURL, log)
	assistantSvc := assistantservice.New(assistantRepo, aiClient, billingSvc, log)

	deps := router.Dependencies{
		Config:    cfg,
		Logger:    log,
		Redis:     redisClient,
		JWT:       jwtManager,
		Auth:      handler.NewAuthHandler(authSvc),
		User:      handler.NewUserHandler(userSvc),
		Journal:   handler.NewJournalHandler(journalSvc),
		Mood:      handler.NewMoodHandler(moodSvc),
		Learning:  handler.NewLearningHandler(learningSvc, billingSvc),
		SafeReach: handler.NewSafeReachHandler(safeReachSvc, redisClient, cfg.HTTP.AllowedOrigins, log),
		Community: handler.NewCommunityHandler(communitySvc),
		Billing:   handler.NewBillingHandler(billingSvc),
		Assistant: handler.NewAssistantHandler(assistantSvc),
		Health:    handler.NewHealthHandler(db, redisClient),
	}

	return &nethttp.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router.New(deps),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}
}

// serveWithGracefulShutdown starts srv and blocks until SIGINT/SIGTERM,
// then gives in-flight requests up to gracePeriod to finish before the
// process exits — so a deploy or restart doesn't cut off a request
// mid-response.
func serveWithGracefulShutdown(srv *nethttp.Server, gracePeriod time.Duration, log zerolog.Logger) error {
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	case sig := <-stop:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		log.Info().Msg("http server shut down cleanly")
	}

	return nil
}
