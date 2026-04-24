package main

import (
	"context"
	"errors"
	"log"
	"net/http"
)

type App struct {
	cfg             Config
	altMount        *AltMountClient
	importProcessor *ImportProcessor
	queueRunner     *QueueRunner
	state           *StateStore
	settings        *SettingsStore
	plex            *PlexClient
	apiServer       *http.Server
}

func NewApp(cfg Config) *App {
	alt := NewAltMountClient(cfg)
	plex := NewPlexClient(cfg)
	state, _ := NewStateStore(cfg.SourceRoot)
	settings, _ := NewSettingsStore(cfg.SourceRoot, cfg)
	proc := NewImportProcessor(cfg, alt, plex, state)
	queue := NewQueueRunner(cfg, proc)
	app := &App{cfg: cfg, altMount: alt, importProcessor: proc, queueRunner: queue, state: state, settings: settings, plex: plex}
	if cfg.HTTPListenAddr != "" {
		app.apiServer = &http.Server{Addr: cfg.HTTPListenAddr, Handler: app.webHandler()}
	}
	return app
}

func (a *App) Run(ctx context.Context) error {
	if a.cfg.SourceRoot == "" {
		return errors.New("WINSTON_SOURCE_ROOT is required")
	}
	if a.apiServer != nil {
		go func() {
			log.Printf("winston: api listening on %s", a.cfg.HTTPListenAddr)
			if err := a.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("winston: api server error: %v", err)
			}
		}()
	}
	if a.cfg.AltMountBaseURL == "" {
		log.Printf("winston: WINSTON_ALTMOUNT_BASE_URL is empty, running in dry bootstrap mode")
		<-ctx.Done()
		return nil
	}
	return a.queueRunner.Run(ctx)
}
