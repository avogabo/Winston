package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type PlexClient struct {
	baseURL  string
	token    string
	pathFrom string
	pathTo   string
	http     *http.Client
}

func NewPlexClient(cfg Config) *PlexClient {
	return &PlexClient{
		baseURL:  strings.TrimRight(cfg.PlexBaseURL, "/"),
		token:    cfg.PlexToken,
		pathFrom: cfg.PlexPathFrom,
		pathTo:   cfg.PlexPathTo,
		http:     &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *PlexClient) RefreshPath(targetPath string) error {
	if p.baseURL == "" || p.token == "" || targetPath == "" {
		return nil
	}
	parent := filepath.Dir(targetPath)
	parent = p.translatePath(parent)
	u := p.baseURL + "/library/sections/all/refresh"
	q := url.Values{}
	q.Set("path", parent)
	q.Set("X-Plex-Token", p.token)
	u += "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("plex refresh failed: %s", resp.Status)
	}
	log.Printf("winston: plex refresh requested for %s", parent)
	return nil
}

func (p *PlexClient) translatePath(in string) string {
	if p.pathFrom == "" || p.pathTo == "" {
		return in
	}
	cleanIn := filepath.Clean(in)
	cleanFrom := filepath.Clean(p.pathFrom)
	cleanTo := filepath.Clean(p.pathTo)
	if cleanIn == cleanFrom {
		return cleanTo
	}
	prefix := cleanFrom + string(filepath.Separator)
	if strings.HasPrefix(cleanIn, prefix) {
		rest := strings.TrimPrefix(cleanIn, prefix)
		return filepath.Join(cleanTo, rest)
	}
	return in
}
