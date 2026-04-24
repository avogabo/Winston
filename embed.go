package main

import "embed"

//go:embed web/dist/* web/dist/assets/*
var webFS embed.FS
