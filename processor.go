package main

import (
	"context"
	"log"
	"path/filepath"
	"time"
)

type ImportProcessor struct {
	cfg Config
	alt *AltMountClient
}

func NewImportProcessor(cfg Config, alt *AltMountClient) *ImportProcessor {
	return &ImportProcessor{cfg: cfg, alt: alt}
}

func (p *ImportProcessor) Run(ctx context.Context) error {
	log.Printf("winston: bootstrap processor ready, source_root=%s", p.cfg.SourceRoot)
	<-ctx.Done()
	return nil
}

func (p *ImportProcessor) ImportOne(ctx context.Context, sourceNZB string) error {
	relativePath := p.buildRelativePath(sourceNZB)
	resp, err := p.alt.ImportFile(ctx, ManualImportRequest{
		FilePath:     sourceNZB,
		RelativePath: relativePath,
	})
	if err != nil {
		return err
	}
	log.Printf("winston: imported source=%s queue_id=%d relative_path=%s", sourceNZB, resp.QueueID, relativePath)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.cfg.SleepBetweenImports):
		return nil
	}
}

func (p *ImportProcessor) buildRelativePath(sourceNZB string) string {
	base := filepath.Base(sourceNZB)
	name := base[:len(base)-len(filepath.Ext(base))]
	return filepath.ToSlash(filepath.Join("pending", name))
}
