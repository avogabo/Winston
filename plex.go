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
	baseURL string
	token   string
	http    *http.Client
}

func NewPlexClient(cfg Config) *PlexClient {
	return &PlexClient{
		baseURL: strings.TrimRight(cfg.PlexBaseURL, "/"),
		token:   cfg.PlexToken,
		http:    &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *PlexClient) RefreshPath(targetPath string) error {
	if p.baseURL == "" || p.token == "" || targetPath == "" {
		return nil
	}
	parent := filepath.Dir(targetPath)
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
