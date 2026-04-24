package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AltMountClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type ManualImportRequest struct {
	FilePath     string `json:"file_path"`
	RelativePath string `json:"relative_path"`
}

type APIEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type ManualImportResponse struct {
	Message string `json:"message"`
	QueueID int    `json:"queue_id"`
}

func NewAltMountClient(cfg Config) *AltMountClient {
	return &AltMountClient{
		baseURL: strings.TrimRight(cfg.AltMountBaseURL, "/"),
		apiKey:  cfg.AltMountAPIKey,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AltMountClient) ImportFile(ctx context.Context, req ManualImportRequest) (*ManualImportResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	u := c.baseURL + "/api/import/file"
	if c.apiKey != "" {
		q := url.Values{}
		q.Set("apikey", c.apiKey)
		u += "?" + q.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("altmount import failed: %s", resp.Status)
	}

	var env APIEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}

	var out ManualImportResponse
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &out); err != nil {
			return nil, err
		}
	}
	return &out, nil
}
