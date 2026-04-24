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
}

func NewApp(cfg Config) *App {
	alt := NewAltMountClient(cfg)
	proc := NewImportProcessor(cfg, alt)
	return &App{cfg: cfg, altMount: alt, importProcessor: proc}
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
	return a.importProcessor.Run(ctx)
}
