package server

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var current *Dungeons

type Dungeons struct {
	Database    *mongo.Database
	MongoClient *mongo.Client
	Router      *gin.Engine

	Version    string
	Port       string
	TokenKey   string
	Origin     string
	LogFormat  string
	Mode       string
	DBHost     string
	DBName     string
	DBTimeout  time.Duration
	TokenTTL   time.Duration
	SeedOnBoot bool
}

func (d *Dungeons) ParseParameters() {
	d.LogFormat = getenv("LOG_FORMAT", "HUMAN")
	d.Version = getenv("API_VERSION", "1.0.0")
	d.Port = normalizePort(getenv("API_PORT", "8080"))
	d.TokenKey = getenv("TOKEN_KEY", "dev-secret")
	d.Origin = getenv("ALLOW_ORIGIN", "*")
	d.Mode = getenv("MODE", "DEVELOP")
	d.DBHost = getenv("DB_HOST", "mongodb://localhost:27017")
	d.DBName = getenv("DB_NAME", "dungeons")
	d.DBTimeout = time.Duration(getenvInt("DB_TIMEOUT_SECONDS", 5)) * time.Second
	d.TokenTTL = time.Duration(getenvInt("TOKEN_TTL_HOURS", 24)) * time.Hour
	d.SeedOnBoot = strings.EqualFold(getenv("SEED_ON_BOOT", "false"), "true")
}

func (d *Dungeons) ListenAndServe() error {
	srv := &http.Server{
		Addr:              d.Port,
		Handler:           d.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Unable to listen and serve")
		return err
	}
	return nil
}

func SetServer(s *Dungeons) {
	current = s
}

func GetServer() *Dungeons {
	return current
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func normalizePort(port string) string {
	port = strings.TrimSpace(port)
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
