package main

import (
	"path/filepath"
	"strings"
)

func (p *ImportProcessor) BuildPreview(sourceNZB string, meta ItemMetadata) *ItemPreview {
	preview := &ItemPreview{
		SourceNZBPath: sourceNZB,
		Metadata:      meta,
		Kind:          normalizeKind(meta.Kind),
		Confidence:    ConfidenceLow,
		State:         StateDetected,
	}

	base := strings.TrimSuffix(filepath.Base(sourceNZB), filepath.Ext(sourceNZB))
	if preview.Kind == "" || preview.Kind == "auto" {
		preview.Kind = guessKindFromPath(sourceNZB)
	}
	if preview.Metadata.Title == "" {
		preview.Metadata.Title = cleanupTitle(base)
	}

	if meta.TMDBID > 0 || meta.TVDBID > 0 || meta.IMDBID != "" {
		preview.Confidence = ConfidenceHigh
		preview.Reason = "explicit_id"
	} else if preview.Metadata.Title != "" && preview.Metadata.Year > 0 {
		preview.Confidence = ConfidenceMedium
		preview.Reason = "title_year_guess"
	} else {
		preview.Confidence = ConfidenceLow
		preview.Reason = "weak_name_parse"
	}

	preview.ProposedPath = p.buildPathForPreview(sourceNZB, preview)
	if preview.Confidence == ConfidenceLow {
		preview.State = StateNeedsReview
		preview.Candidates = []CandidateMatch{
			{Label: preview.Metadata.Title, Kind: preview.Kind, TMDBID: preview.Metadata.TMDBID, TVDBID: preview.Metadata.TVDBID, IMDBID: preview.Metadata.IMDBID, Year: preview.Metadata.Year, Reason: "current_guess", Score: 50},
		}
	}
	return preview
}

func (p *ImportProcessor) buildPathForPreview(sourceNZB string, preview *ItemPreview) string {
	if preview.Metadata.RelativePathOverride != "" {
		return preview.Metadata.RelativePathOverride
	}
	return p.buildRelativePath(sourceNZB)
}

func normalizeKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "movie", "series", "episode", "auto":
		return k
	default:
		return "auto"
	}
}

func guessKindFromPath(path string) string {
	low := strings.ToLower(path)
	if strings.Contains(low, "season") || strings.Contains(low, "temporada") || strings.Contains(low, "series") {
		return "series"
	}
	return "movie"
}

func cleanupTitle(s string) string {
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}
