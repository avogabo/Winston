package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

type ReviewListItem struct {
	SourceNZBPath string           `json:"source_nzb_path"`
	State         ItemState        `json:"state"`
	Confidence    MatchConfidence  `json:"confidence"`
	Metadata      ItemMetadata     `json:"metadata"`
	ProposedPath  string           `json:"proposed_path"`
	Reason        string           `json:"reason"`
	Candidates    []CandidateMatch `json:"candidates,omitempty"`
	QueueID       int              `json:"queue_id,omitempty"`
	Status        string           `json:"status,omitempty"`
}

type ReviewListResponse struct {
	Items []ReviewListItem `json:"items"`
}

type ApplyCorrectionRequest struct {
	TMDBID               *int   `json:"tmdb_id,omitempty"`
	RelativePathOverride string `json:"relative_path_override,omitempty"`
}

type ApproveImportRequest struct {
	Action string `json:"action"`
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/review/items", a.handleReviewItems)
	mux.HandleFunc("/api/review/item", a.handleReviewItem)
	mux.HandleFunc("/api/review/correct", a.handleReviewCorrect)
	mux.HandleFunc("/api/settings", a.handleSettings)
	mux.HandleFunc("/api/filebot/status", a.handleFileBotStatus)
	mux.HandleFunc("/api/review/approve", a.handleReviewApprove)
	mux.HandleFunc("/api/review/import", a.handleReviewImport)
	mux.HandleFunc("/api/review/reset", a.handleReviewReset)
	return mux
}

func (a *App) handleReviewReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		http.Error(w, "missing source", http.StatusBadRequest)
		return
	}
	if a.state == nil {
		http.Error(w, "state store unavailable", http.StatusInternalServerError)
		return
	}
	if err := a.state.Delete(source); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	preview := a.importProcessor.BuildPreview(source, ItemMetadata{})
	writeJSON(w, http.StatusOK, preview)
}

func (a *App) handleFileBotStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := a.effectiveConfig()
	status := NewFileBotClient(cfg).Status(r.Context())
	writeJSON(w, http.StatusOK, status)
}

func (a *App) handleReviewApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		http.Error(w, "missing source", http.StatusBadRequest)
		return
	}
	preview, err := a.importProcessor.Approve(source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (a *App) handleReviewImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		http.Error(w, "missing source", http.StatusBadRequest)
		return
	}
	if err := a.importApproved(r.Context(), source); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item, ok := a.reviewItem(source)
	if !ok {
		http.Error(w, "not found after import", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	if a.settings == nil {
		http.Error(w, "settings store unavailable", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, a.settings.Get())
	case http.MethodPost:
		var req Settings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := a.settings.Put(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) handleReviewItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items := a.reviewItems()
	writeJSON(w, http.StatusOK, ReviewListResponse{Items: items})
}

func (a *App) handleReviewItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		http.Error(w, "missing source", http.StatusBadRequest)
		return
	}
	item, ok := a.reviewItem(source)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleReviewCorrect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		http.Error(w, "missing source", http.StatusBadRequest)
		return
	}
	var req ApplyCorrectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var (
		preview *ItemPreview
		err     error
	)
	switch {
	case req.TMDBID != nil:
		preview, err = a.importProcessor.ApplyTMDBCorrection(source, *req.TMDBID)
	case strings.TrimSpace(req.RelativePathOverride) != "":
		preview, err = a.importProcessor.ApplyRelativePathOverride(source, strings.TrimSpace(req.RelativePathOverride))
	default:
		http.Error(w, "no correction provided", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (a *App) reviewItems() []ReviewListItem {
	if a.state == nil {
		return nil
	}
	items := make([]ReviewListItem, 0, len(a.state.Data.Imported))
	for source, rec := range a.state.Data.Imported {
		items = append(items, recordToReviewItem(source, rec))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].State != items[j].State {
			return items[i].State < items[j].State
		}
		return items[i].SourceNZBPath < items[j].SourceNZBPath
	})
	return items
}

func (a *App) reviewItem(source string) (ReviewListItem, bool) {
	if a.state == nil {
		return ReviewListItem{}, false
	}
	rec, ok := a.state.Data.Imported[source]
	if !ok {
		return ReviewListItem{}, false
	}
	return recordToReviewItem(source, rec), true
}

func recordToReviewItem(source string, rec ImportedRecord) ReviewListItem {
	item := ReviewListItem{
		SourceNZBPath: source,
		State:         rec.State,
		Confidence:    rec.Confidence,
		Metadata:      rec.Metadata,
		ProposedPath:  rec.RelativePath,
		QueueID:       rec.QueueID,
		Status:        rec.Status,
	}
	if rec.Preview != nil {
		item.ProposedPath = rec.Preview.ProposedPath
		item.Reason = rec.Preview.Reason
		item.Candidates = rec.Preview.Candidates
		if item.Confidence == "" {
			item.Confidence = rec.Preview.Confidence
		}
		if item.State == "" {
			item.State = rec.Preview.State
		}
		if item.Metadata == (ItemMetadata{}) {
			item.Metadata = rec.Preview.Metadata
		}
	}
	return item
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
