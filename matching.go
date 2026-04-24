package main

import "strings"

type Matcher struct{}

func NewMatcher() *Matcher { return &Matcher{} }

func (m *Matcher) Resolve(meta ItemMetadata, sourceNZB string) (ItemMetadata, MatchConfidence, []CandidateMatch, string) {
	if meta.RelativePathOverride != "" {
		return meta, ConfidenceHigh, nil, "relative_path_override"
	}
	if meta.TMDBID > 0 || meta.TVDBID > 0 || meta.IMDBID != "" {
		return meta, ConfidenceHigh, nil, "explicit_id"
	}
	if meta.Title != "" && meta.Year > 0 {
		return meta, ConfidenceMedium, nil, "manual_title_year"
	}

	base := cleanupTitle(strings.TrimSuffix(filepathBase(sourceNZB), filepathExt(sourceNZB)))
	meta.Title = base
	return meta, ConfidenceLow, []CandidateMatch{{Label: base, Kind: normalizeKind(meta.Kind), Reason: "name_parse", Score: 40}}, "name_parse"
}

func filepathBase(path string) string {
	parts := strings.Split(strings.ReplaceAll(path, "\\", "/"), "/")
	if len(parts) == 0 {
		return path
	}
	return parts[len(parts)-1]
}

func filepathExt(path string) string {
	base := filepathBase(path)
	idx := strings.LastIndex(base, ".")
	if idx < 0 {
		return ""
	}
	return base[idx:]
}
