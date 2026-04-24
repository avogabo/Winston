package main

import (
	"context"
	"log"
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

	preview := p.BuildPreview(sourceNZB, ItemMetadata{})
	if preview.State == StateNeedsReview {
		log.Printf("winston: review required for %s proposed=%s reason=%s", sourceNZB, preview.ProposedPath, preview.Reason)
		if p.state != nil {
			_ = p.state.Put(sourceNZB, ImportedRecord{RelativePath: preview.ProposedPath, Status: "review", State: preview.State, Confidence: preview.Confidence, Metadata: preview.Metadata, Preview: preview})
		}
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
		_ = p.state.Put(sourceNZB, ImportedRecord{QueueID: resp.QueueID, RelativePath: relativePath, Status: "submitted", State: StateImporting, Confidence: preview.Confidence, Metadata: preview.Metadata, Preview: preview})
	}

	item, err := p.waitForQueueReady(ctx, resp.QueueID)
	if err != nil {
		if p.state != nil {
			_ = p.state.Put(sourceNZB, ImportedRecord{QueueID: resp.QueueID, RelativePath: relativePath, Status: "error", State: StateFailed, Confidence: preview.Confidence, Metadata: preview.Metadata, Preview: preview})
		}
		return err
	}
	if p.state != nil {
		_ = p.state.Put(sourceNZB, ImportedRecord{QueueID: resp.QueueID, RelativePath: item.TargetPath, Status: item.Status, State: StateImported, Confidence: preview.Confidence, Metadata: preview.Metadata, Preview: preview})
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.cfg.SleepBetweenImports):
		return nil
	}
}
