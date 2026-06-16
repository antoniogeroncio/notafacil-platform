// Command api is the NotaFácil backend HTTP server.
//
// Dependency injection is wired here (Princípio I): concrete implementations
// (Mongo repositories, SMTP/Fake e-mail sender, token manager) are constructed
// and injected into services and handlers.
package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/notafacil/platform/backend/internal/auth"
	"github.com/notafacil/platform/backend/internal/invite"
	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/config"
	appdb "github.com/notafacil/platform/backend/pkg/db"
	"github.com/notafacil/platform/backend/pkg/email"
	"github.com/notafacil/platform/backend/pkg/httpx"
	"github.com/notafacil/platform/backend/pkg/middleware"
	"github.com/notafacil/platform/backend/pkg/token"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	database, client, err := appdb.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()
	if err := appdb.EnsureIndexes(ctx, database); err != nil {
		log.Fatalf("mongo indexes: %v", err)
	}

	deps := buildDeps(cfg)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		mountRoutes(api, database, deps)
	})

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// deps holds the wired dependencies shared across handlers.
type deps struct {
	cfg          config.Config
	tokenManager *token.Manager
	emailSender  email.Sender
}

func buildDeps(cfg config.Config) deps {
	var sender email.Sender
	if cfg.SMTP.Host != "" {
		sender = email.NewSMTPSender(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.From)
	} else {
		// No SMTP configured (dev): capture e-mails in memory.
		sender = &email.FakeSender{}
	}
	return deps{
		cfg:          cfg,
		tokenManager: token.NewManager(cfg.JWTSecret, cfg.SessionTTL),
		emailSender:  sender,
	}
}

// mountRoutes wires repositories, services and handlers (DI) and installs the
// domain routes.
func mountRoutes(api chi.Router, database *mongo.Database, deps deps) {
	userRepo := user.NewMongoRepository(database)
	inviteRepo := invite.NewMongoRepository(database)

	inviteSvc := invite.NewService(userRepo, inviteRepo, deps.emailSender, deps.cfg.InviteTTL, deps.cfg.AppBaseURL)
	inviteHandler := invite.NewHandler(inviteSvc)

	authSvc := auth.NewService(userRepo, inviteRepo, deps.tokenManager)
	secureCookies := strings.HasPrefix(deps.cfg.AppBaseURL, "https://")
	authHandler := auth.NewHandler(authSvc, deps.cfg.SessionTTL, secureCookies)

	// Public routes (no session required).
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)
	api.Post("/invites/{token}/accept", authHandler.AcceptInvite)

	// Authenticated routes: identity (tenant + role) derived from the session.
	api.Group(func(authed chi.Router) {
		authed.Use(middleware.Authenticate(deps.tokenManager))

		authed.Get("/me", authHandler.Me)

		authed.With(middleware.RequireRole(string(user.RoleAdmin))).
			Post("/invites", inviteHandler.Create)
	})
}
