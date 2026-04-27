package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type SettingsStore struct {
	path string
	mu   sync.Mutex
	data Settings
}

type Settings struct {
	AltMountBaseURL     string `json:"altmount_base_url"`
	AltMountAPIKey      string `json:"altmount_api_key"`
	AltMountPathFrom    string `json:"altmount_path_from"`
	AltMountPathTo      string `json:"altmount_path_to"`
	PlexBaseURL         string `json:"plex_base_url"`
	PlexToken           string `json:"plex_token"`
	PlexPathFrom        string `json:"plex_path_from"`
	PlexPathTo          string `json:"plex_path_to"`
	DefaultMode         string `json:"default_mode"`
	SleepBetweenImports string `json:"sleep_between_imports"`
	MoviesTemplate      string `json:"movies_template"`
	SeriesTemplate      string `json:"series_template"`
	FileBotMovieFormat  string `json:"filebot_movie_format"`
	FileBotSeriesFormat string `json:"filebot_series_format"`
	FileBotDB           string `json:"filebot_db"`
	FileBotBinary       string `json:"filebot_binary"`
	FileBotHome         string `json:"filebot_home"`
	AutoImportMedium    bool   `json:"auto_import_medium"`
}

func NewSettingsStore(root string, cfg Config) (*SettingsStore, error) {
	path := filepath.Join(root, ".winston-settings.json")
	s := &SettingsStore{path: path, data: Settings{
		AltMountBaseURL:     cfg.AltMountBaseURL,
		AltMountAPIKey:      cfg.AltMountAPIKey,
		AltMountPathFrom:    cfg.AltMountPathFrom,
		AltMountPathTo:      cfg.AltMountPathTo,
		PlexBaseURL:         cfg.PlexBaseURL,
		PlexToken:           cfg.PlexToken,
		PlexPathFrom:        cfg.PlexPathFrom,
		PlexPathTo:          cfg.PlexPathTo,
		DefaultMode:         cfg.DefaultMode,
		SleepBetweenImports: cfg.SleepBetweenImports.String(),
		MoviesTemplate:      cfg.MoviesTemplate,
		SeriesTemplate:      cfg.SeriesTemplate,
		FileBotMovieFormat:  cfg.FileBotMovieFormat,
		FileBotSeriesFormat: cfg.FileBotSeriesFormat,
		FileBotDB:           cfg.FileBotDB,
		FileBotBinary:       cfg.FileBotBinary,
		FileBotHome:         cfg.FileBotHome,
		AutoImportMedium:    cfg.AutoImportMedium,
	}}
	b, err := os.ReadFile(path)
	if err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &s.data)
	}
	return s, nil
}

func (s *SettingsStore) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *SettingsStore) Put(v Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = v
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}
