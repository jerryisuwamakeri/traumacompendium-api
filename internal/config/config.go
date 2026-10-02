// Package config loads and validates the application's runtime configuration
// from environment variables (optionally backed by a local .env file).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Mode switches an integration between a safe local/dev behavior and the
// real, live behavior. It is used for every outbound third-party
// integration (SMS, email) so the app never has to talk to a paid provider
// during development, testing, or CI.
type Mode string

const (
	ModeSandbox    Mode = "sandbox"
	ModeProduction Mode = "production"
)

// Config is the fully-resolved application configuration, grouped by
// subsystem. It is constructed once at startup via Load and passed down
// explicitly to every component that needs it — no global state.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OTP      OTPConfig
	Termii   TermiiConfig
	SMTP     SMTPConfig
	TTS      TTSConfig
	HTTP     HTTPConfig
}

type AppConfig struct {
	Env      string // development | staging | production
	Port     string
	BaseURL  string
	LogLevel string
}

func (a AppConfig) IsProduction() bool { return a.Env == "production" }

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN builds a go-sql-driver/mysql compatible data source name.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
}

type OTPConfig struct {
	Length      int
	TTL         time.Duration
	MaxAttempts int
}

// TermiiConfig configures the Termii SMS/OTP integration. Mode gates
// whether the client performs real HTTP calls against Termii's API
// (production) or simply logs what would have been sent (sandbox).
type TermiiConfig struct {
	Mode     Mode
	BaseURL  string
	APIKey   string
	SenderID string
}

// SMTPConfig configures outbound transactional email. Mode mirrors
// TermiiConfig.Mode: sandbox logs the message instead of dialing out.
type SMTPConfig struct {
	Mode     Mode
	Host     string
	Port     string
	Username string
	Password string
	FromAddr string
	FromName string
}

// TTSConfig configures the lesson-audio text-to-speech integration.
// Mode mirrors TermiiConfig.Mode: sandbox returns a placeholder clip
// instead of calling a real provider.
type TTSConfig struct {
	Mode    Mode
	BaseURL string
	APIKey  string
}

type HTTPConfig struct {
	AllowedOrigins         []string
	RateLimitPerMinute     int
	AuthRateLimitPerMinute int
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	ShutdownGracePeriod    time.Duration
	IdempotencyTTL         time.Duration
}

// Load reads configuration from the process environment. If a .env file is
// present in the working directory it is loaded first (local/dev
// convenience); in containers and CI, real environment variables are used
// and no .env file is expected to exist.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Port:     getEnv("APP_PORT", "8080"),
			BaseURL:  getEnv("APP_BASE_URL", "http://localhost:8080"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			User:            getEnv("DB_USER", "trauma_app"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "trauma_compendium"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:     getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvAsDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
			Issuer:        getEnv("JWT_ISSUER", "traumacompendium-api"),
		},
		OTP: OTPConfig{
			Length:      getEnvAsInt("OTP_LENGTH", 6),
			TTL:         getEnvAsDuration("OTP_TTL", 10*time.Minute),
			MaxAttempts: getEnvAsInt("OTP_MAX_ATTEMPTS", 5),
		},
		Termii: TermiiConfig{
			Mode:     Mode(getEnv("TERMII_MODE", string(ModeSandbox))),
			BaseURL:  getEnv("TERMII_BASE_URL", "https://api.ng.termii.com"),
			APIKey:   getEnv("TERMII_API_KEY", ""),
			SenderID: getEnv("TERMII_SENDER_ID", "TraumaComp"),
		},
		SMTP: SMTPConfig{
			Mode:     Mode(getEnv("SMTP_MODE", string(ModeSandbox))),
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			FromAddr: getEnv("SMTP_FROM_ADDRESS", "no-reply@traumacompendium.app"),
			FromName: getEnv("SMTP_FROM_NAME", "TraumaCompendium"),
		},
		TTS: TTSConfig{
			Mode:    Mode(getEnv("TTS_MODE", string(ModeSandbox))),
			BaseURL: getEnv("TTS_BASE_URL", "https://texttospeech.googleapis.com"),
			APIKey:  getEnv("TTS_API_KEY", ""),
		},
		HTTP: HTTPConfig{
			AllowedOrigins:         splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
			RateLimitPerMinute:     getEnvAsInt("RATE_LIMIT_PER_MINUTE", 120),
			AuthRateLimitPerMinute: getEnvAsInt("AUTH_RATE_LIMIT_PER_MINUTE", 10),
			ReadTimeout:            getEnvAsDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:           getEnvAsDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			ShutdownGracePeriod:    getEnvAsDuration("HTTP_SHUTDOWN_GRACE_PERIOD", 15*time.Second),
			IdempotencyTTL:         getEnvAsDuration("IDEMPOTENCY_TTL", 24*time.Hour),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate enforces the invariants that must hold before the app is allowed
// to start: secrets must be non-empty, and in production the sandbox
// integrations must not be silently in use.
func (c *Config) validate() error {
	var missing []string

	if c.JWT.AccessSecret == "" {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}
	if c.JWT.RefreshSecret == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}
	if c.JWT.AccessSecret != "" && c.JWT.AccessSecret == c.JWT.RefreshSecret {
		return fmt.Errorf("config: JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different")
	}

	if c.App.IsProduction() {
		if c.Database.Password == "" {
			missing = append(missing, "DB_PASSWORD")
		}
		if c.Termii.Mode == ModeProduction && c.Termii.APIKey == "" {
			missing = append(missing, "TERMII_API_KEY")
		}
		if c.SMTP.Mode == ModeProduction && (c.SMTP.Host == "" || c.SMTP.Username == "") {
			missing = append(missing, "SMTP_HOST/SMTP_USERNAME")
		}
		if c.TTS.Mode == ModeProduction && c.TTS.APIKey == "" {
			missing = append(missing, "TTS_API_KEY")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("config: missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitAndTrim(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
