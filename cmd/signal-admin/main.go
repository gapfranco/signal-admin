package main

import (
	"context"
	"html/template"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"signal-admin/config"
	"signal-admin/internal/migrations"
	"signal-admin/internal/storage"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

type application struct {
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
	db             *storage.TursoDB
	config         config.Config
	setupDone      atomic.Bool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := storage.NewTursoDB(storage.DBConfig{
		URL:       cfg.DBURL,
		Token:     cfg.DBToken,
		Mode:      cfg.DBMode,
		LocalPath: cfg.DBLocalPath,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Run(db.DB()); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if cfg.DBMode == "sync" {
		if err := migrations.RunRemote(cfg.DBURL, cfg.DBToken); err != nil {
			logger.Warn("failed to run remote migrations", "error", err)
		}
	}

	if err := db.SyncStartup(context.Background()); err != nil {
		logger.Error("db sync after migrations failed", "error", err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("failed to create template cache", "error", err)
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	app := &application{
		logger:         logger,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    form.NewDecoder(),
		db:             db,
		config:         cfg,
	}

	hasUsers, err := db.HasUsers()
	if err != nil {
		logger.Error("failed to check users", "error", err)
		os.Exit(1)
	}
	app.setupDone.Store(hasUsers)

	runDesktop(app, cfg, logger)
}
