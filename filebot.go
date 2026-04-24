package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type FileBotClient struct {
	cfg Config
}

type FileBotResult struct {
	Output string `json:"output"`
}

func NewFileBotClient(cfg Config) *FileBotClient {
	return &FileBotClient{cfg: cfg}
}

func (f *FileBotClient) Enabled() bool {
	return strings.EqualFold(f.cfg.DefaultMode, "filebot")
}

func (f *FileBotClient) Resolve(ctx context.Context, sourceNZB string, meta ItemMetadata) (string, string, error) {
	if !f.Enabled() {
		return "", "", nil
	}
	format := f.cfg.FileBotMovieFormat
	if normalizeKind(meta.Kind) == "series" {
		format = f.cfg.FileBotSeriesFormat
	}
	if strings.TrimSpace(format) == "" {
		return "", "", nil
	}

	payload := map[string]any{
		"source":  sourceNZB,
		"title":   meta.Title,
		"year":    meta.Year,
		"season":  meta.Season,
		"episode": meta.Episode,
		"tmdb_id": meta.TMDBID,
		"tvdb_id": meta.TVDBID,
		"imdb_id": meta.IMDBID,
		"kind":    meta.Kind,
		"format":  format,
		"db":      f.cfg.FileBotDB,
	}
	b, _ := json.Marshal(payload)

	py := `
import json, os, re, sys
p = json.loads(sys.stdin.read())
fmt = p.get("format") or ""
title = p.get("title") or "Unknown"
year = p.get("year") or 0
season = p.get("season") or 0
episode = p.get("episode") or 0
kind = (p.get("kind") or "movie").lower()
quality = "1080p"
alpha = title[:1].upper() if title else "#"
series = title
mapping = {
  "title": title,
  "year": str(year) if year else "",
  "season": f"{season:02d}" if season else "01",
  "episode": f"{season:02d}x{episode:02d}" if episode else "01x01",
  "quality": quality,
  "alpha": alpha,
  "series": series,
  "plex": title,
}
out = fmt
for k, v in mapping.items():
  out = out.replace("{" + k + "}", str(v))
out = re.sub(r'/+', '/', out).strip('/ ')
print(out)
`
	cmd := exec.CommandContext(ctx, "python3", "-c", py)
	cmd.Stdin = bytes.NewReader(b)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", stderr.String(), fmt.Errorf("filebot resolver failed: %w", err)
	}
	resolved := strings.TrimSpace(stdout.String())
	if resolved == "" {
		return "", stderr.String(), nil
	}
	return resolved, stderr.String(), nil
}

func (f *FileBotClient) Probe(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "filebot", "-version")
	return cmd.Run()
}
