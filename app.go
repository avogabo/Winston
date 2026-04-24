package main

import (
	"context"
	"errors"
	"log"
)

type App struct {
	cfg             Config
	altMount        *AltMountClient
	importProcessor *ImportProcessor
	queueRunner     *QueueRunner
	state           *StateStore
	plex            *PlexClient
}

func NewApp(cfg Config) *App {
	alt := NewAltMountClient(cfg)
	plex := NewPlexClient(cfg)
	state, _ := NewStateStore(cfg.SourceRoot)
	proc := NewImportProcessor(cfg, alt, plex, state)
	queue := NewQueueRunner(cfg, proc)
	return &App{cfg: cfg, altMount: alt, importProcessor: proc, queueRunner: queue, state: state, plex: plex}
}

func (a *App) Run(ctx context.Context) error {
	if a.cfg.AltMountBaseURL == "" {
		log.Printf("winston: WINSTON_ALTMOUNT_BASE_URL is empty, running in dry bootstrap mode")
		<-ctx.Done()
		return nil
	}
	if a.cfg.SourceRoot == "" {
		return errors.New("WINSTON_SOURCE_ROOT is required")
	}
	return a.queueRunner.Run(ctx)
}
