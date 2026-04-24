package main

import (
	"context"
	"log"
	"path/filepath"
	"time"
)

type ImportProcessor struct {
	cfg   Config
	alt   *AltMountClient
	state *StateStore
}

func NewImportProcessor(cfg Config, alt *AltMountClient, state *StateStore) *ImportProcessor {
	return &ImportProcessor{cfg: cfg, alt: alt, state: state}
}

func (p *ImportProcessor) Run(ctx context.Context) error {
	log.Printf("winston: bootstrap processor ready, source_root=%s", p.cfg.SourceRoot)
	<-ctx.Done()
	return nil
}

func (p *ImportProcessor) ImportOne(ctx context.Context, sourceNZB string) error {
	relativePath := p.buildRelativePath(sourceNZB)
	if p.state != nil && p.state.Has(sourceNZB) {
		log.Printf("winston: skipping already imported nzb: %s", sourceNZB)
		return nil
	}

	resp, err := p.alt.ImportFile(ctx, ManualImportRequest{
		FilePath:     sourceNZB,
		RelativePath: relativePath,
	})
	if err != nil {
		return err
	}
	log.Printf("winston: imported source=%s queue_id=%d relative_path=%s", sourceNZB, resp.QueueID, relativePath)
	if p.state != nil {
		_ = p.state.Put(sourceNZB, ImportedRecord{QueueID: resp.QueueID, RelativePath: relativePath, Status: "submitted"})
	}

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
