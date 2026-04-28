package main

import (
	"context"
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
	if strings.TrimSpace(preview.Metadata.Quality) == "" {
		preview.Metadata.Quality = detectQualityForSource(sourceNZB)
	}

	resolved, confidence, candidates, reason := p.matcher.Resolve(preview.Metadata, sourceNZB)
	preview.Metadata = resolved
	preview.Confidence = confidence
	preview.Reason = reason
	preview.Candidates = candidates

	preview.ProposedPath = p.buildPathForPreview(sourceNZB, preview)
	if preview.Confidence == ConfidenceLow {
		preview.State = StateNeedsReview
	} else {
		preview.State = StateApproved
	}
	return preview
}

func (p *ImportProcessor) buildPathForPreview(sourceNZB string, preview *ItemPreview) string {
	if preview.Metadata.RelativePathOverride != "" {
		return preview.Metadata.RelativePathOverride
	}
	if p.filebot != nil && p.filebot.Enabled() {
		if resolved, err := p.filebot.Resolve(context.Background(), sourceNZB, preview.Metadata); err == nil && resolved != nil && strings.TrimSpace(resolved.RelativePath) != "" {
			if preview.Metadata.ResolvedEpisodeTitle == "" && strings.TrimSpace(resolved.EpisodeTitle) != "" {
				preview.Metadata.ResolvedEpisodeTitle = strings.TrimSpace(resolved.EpisodeTitle)
			}
			return resolved.RelativePath
		}
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
