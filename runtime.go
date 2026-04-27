package main

import (
	"context"
	"fmt"
	"time"
)

func (a *App) effectiveConfig() Config {
	cfg := a.cfg
	if a.settings == nil {
		return cfg
	}
	return applySettingsToConfig(cfg, a.settings.Get())
}

func applySettingsToConfig(cfg Config, s Settings) Config {
	if s.AltMountBaseURL != "" {
		cfg.AltMountBaseURL = s.AltMountBaseURL
	}
	if s.AltMountAPIKey != "" {
		cfg.AltMountAPIKey = s.AltMountAPIKey
	}
	if s.AltMountPathFrom != "" {
		cfg.AltMountPathFrom = s.AltMountPathFrom
	}
	if s.AltMountPathTo != "" {
		cfg.AltMountPathTo = s.AltMountPathTo
	}
	if s.PlexBaseURL != "" {
		cfg.PlexBaseURL = s.PlexBaseURL
	}
	if s.PlexToken != "" {
		cfg.PlexToken = s.PlexToken
	}
	if s.PlexPathFrom != "" {
		cfg.PlexPathFrom = s.PlexPathFrom
	}
	if s.PlexPathTo != "" {
		cfg.PlexPathTo = s.PlexPathTo
	}
	if s.DefaultMode != "" {
		cfg.DefaultMode = s.DefaultMode
	}
	if s.MoviesTemplate != "" {
		cfg.MoviesTemplate = s.MoviesTemplate
	}
	if s.SeriesTemplate != "" {
		cfg.SeriesTemplate = s.SeriesTemplate
	}
	if s.FileBotMovieFormat != "" {
		cfg.FileBotMovieFormat = s.FileBotMovieFormat
	}
	if s.FileBotSeriesFormat != "" {
		cfg.FileBotSeriesFormat = s.FileBotSeriesFormat
	}
	if s.FileBotDB != "" {
		cfg.FileBotDB = s.FileBotDB
	}
	if s.FileBotBinary != "" {
		cfg.FileBotBinary = s.FileBotBinary
	}
	if s.SleepBetweenImports != "" {
		if d, err := time.ParseDuration(s.SleepBetweenImports); err == nil {
			cfg.SleepBetweenImports = d
		}
	}
	cfg.AutoImportMedium = s.AutoImportMedium
	return cfg
}

func (p *ImportProcessor) applyRuntimeConfig(cfg Config) {
	p.cfg = cfg
	p.alt = NewAltMountClient(cfg)
	p.plex = NewPlexClient(cfg)
	p.filebot = NewFileBotClient(cfg)
}

func (a *App) importApproved(ctx context.Context, source string) error {
	if a.importProcessor == nil {
		return fmt.Errorf("processor unavailable")
	}
	a.importProcessor.applyRuntimeConfig(a.effectiveConfig())
	return a.importProcessor.ImportOne(ctx, source)
}
