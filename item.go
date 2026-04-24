package main

type ItemMetadata struct {
	TMDBID               int    `json:"tmdb_id"`
	TVDBID               int    `json:"tvdb_id"`
	IMDBID               string `json:"imdb_id"`
	Kind                 string `json:"kind"`
	Title                string `json:"title"`
	Year                 int    `json:"year"`
	Season               int    `json:"season"`
	Episode              int    `json:"episode"`
	Quality              string `json:"quality"`
	RelativePathOverride string `json:"relative_path_override"`
}
