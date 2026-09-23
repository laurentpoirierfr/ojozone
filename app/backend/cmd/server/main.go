package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	ojozone "github.com/laurentpoirierfr/ojozone"
	_ "github.com/laurentpoirierfr/ojozone/docs"
	"github.com/laurentpoirierfr/ojozone/internal/auth"
	"github.com/laurentpoirierfr/ojozone/internal/httpapi"
	"github.com/laurentpoirierfr/ojozone/internal/repository"
	"github.com/laurentpoirierfr/ojozone/internal/service"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// @title OjoZone API
// @version 1.0
// @description API de comparaison du coût de la vie dans la zone euro.
// @description Les montants décimaux sont exposés sous forme de chaînes pour préserver leur précision.
// @BasePath /
// @schemes http https
// @produce json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	pool, err := pgxpool.New(context.Background(), getenv("DATABASE_URL", "postgres://ojozone@localhost:5432/ojozone?sslmode=disable&search_path=public"))
	if err != nil {
		slog.Error("configuration PostgreSQL invalide", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	staticFS, err := ojozone.StaticFS()
	if err != nil {
		slog.Error("chargement des fichiers statiques impossible", "error", err)
		os.Exit(1)
	}

	repository := repository.NewPostgreSQL(db.New(pool), pool)
	tokenManager := auth.NewManager(authSecret(), getenv("AUTH_ISSUER", "ojozone-api"), accessTokensTTL())
	appService := service.New(repository, tokenManager)
	router := httpapi.NewRouter(appService, tokenManager, staticFS, httpapi.BuildInfo{
		Name: "ojozone-api", Version: version, Commit: commit, BuildTime: buildTime,
	})

	server := &http.Server{
		Addr:              ":" + getenv("PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("serveur OjoZone démarré", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signal := <-signals:
		slog.Info("arrêt demandé", "signal", signal.String())
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("arrêt inattendu du serveur", "error", err)
			os.Exit(1)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		slog.Error("arrêt propre impossible", "error", err)
		os.Exit(1)
	}
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

// authSecret retourne la clé de signature des jetons, ou génère une clé éphémère en développement.
func authSecret() string {
	secret := getenv("AUTH_JWT_SECRET", "")
	if secret == "" {
		nonce := make([]byte, 32)
		_, _ = rand.Read(nonce)
		secret = base64.RawURLEncoding.EncodeToString(nonce)
		slog.Warn("AUTH_JWT_SECRET absent : clé éphémère générée, les sessions seront invalidées au redémarrage")
	}
	if len(secret) < 32 {
		slog.Warn("AUTH_JWT_SECRET trop courte, une longueur d'au moins 32 octets est recommandée")
	}
	return secret
}

// accessTokensTTL parse la durée de vie des jetons d'accès, 15 minutes par défaut.
func accessTokensTTL() time.Duration {
	ttl, err := time.ParseDuration(getenv("AUTH_ACCESS_TTL", "15m"))
	if err != nil || ttl <= 0 {
		return 15 * time.Minute
	}
	return ttl
}
