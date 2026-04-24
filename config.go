package main

import (
	"os"
	"time"
)

type Config struct {
	HTTPListenAddr      string
	AltMountBaseURL     string
	AltMountAPIKey      string
	PlexBaseURL         string
	PlexToken           string
	SourceRoot          string
	DefaultMode         string
	MoviesTemplate      string
	SeriesTemplate      string
	FileBotMovieFormat  string
	FileBotSeriesFormat string
	FileBotDB           string
	FileBotHome         string
	PlexPathFrom        string
	PlexPathTo          string
	SleepBetweenImports time.Duration
	AutoImportMedium    bool
}

func LoadConfig() Config {
	sleep := 3 * time.Second
	if raw := os.Getenv("WINSTON_SLEEP_BETWEEN_IMPORTS"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			sleep = d
		}
	}

	mode := getenv("WINSTON_DEFAULT_MODE", "preserve")
	autoImportMedium := getenv("WINSTON_AUTOIMPORT_MEDIUM", "true") != "false"

	return Config{
		HTTPListenAddr:      getenv("WINSTON_HTTP_LISTEN_ADDR", ":8091"),
		AltMountBaseURL:     os.Getenv("WINSTON_ALTMOUNT_BASE_URL"),
		AltMountAPIKey:      os.Getenv("WINSTON_ALTMOUNT_API_KEY"),
		PlexBaseURL:         os.Getenv("WINSTON_PLEX_BASE_URL"),
		PlexToken:           os.Getenv("WINSTON_PLEX_TOKEN"),
		SourceRoot:          os.Getenv("WINSTON_SOURCE_ROOT"),
		DefaultMode:         mode,
		MoviesTemplate:      getenv("WINSTON_MOVIES_TEMPLATE", "Peliculas/{quality}/{alpha}/{title} ({year})"),
		SeriesTemplate:      getenv("WINSTON_SERIES_TEMPLATE", "Series/{alpha}/{series}/Temporada {season}/{series} - {episode}"),
		FileBotMovieFormat:  getenv("WINSTON_FILEBOT_FORMAT_MOVIE", "Peliculas/{plex}"),
		FileBotSeriesFormat: getenv("WINSTON_FILEBOT_FORMAT_SERIES", "Series/{plex}"),
		FileBotDB:           getenv("WINSTON_FILEBOT_DB", "TheMovieDB"),
		FileBotHome:         getenv("FILEBOT_HOME", "/config/filebot"),
		PlexPathFrom:        getenv("WINSTON_PLEX_PATH_FROM", ""),
		PlexPathTo:          getenv("WINSTON_PLEX_PATH_TO", ""),
		SleepBetweenImports: sleep,
		AutoImportMedium:    autoImportMedium,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
