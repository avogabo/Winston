package main

import (
	"context"
	"log"
	"time"
)

type ImportProcessor struct {
	cfg     Config
	alt     *AltMountClient
	plex    *PlexClient
	state   *StateStore
	matcher *Matcher
}

func NewImportProcessor(cfg Config, alt *AltMountClient, plex *PlexClient, state *StateStore) *ImportProcessor {
	return &ImportProcessor{cfg: cfg, alt: alt, plex: plex, state: state, matcher: NewMatcher()}
}

func (p *ImportProcessor) Run(ctx context.Context) error {
	log.Printf("winston: bootstrap processor ready, source_root=%s", p.cfg.SourceRoot)
	<-ctx.Done()
	return nil
}

func (p *ImportProcessor) ImportOne(ctx context.Context, sourceNZB string) error {
	preview := p.BuildPreview(sourceNZB, ItemMetadata{})
	if p.state != nil {
		if rec, ok := p.state.Data.Imported[sourceNZB]; ok && rec.Metadata != (ItemMetadata{}) {
			preview = p.BuildPreview(sourceNZB, rec.Metadata)
		}
	}
	relativePath := preview.ProposedPath
	if p.state != nil && p.state.Has(sourceNZB) {
		log.Printf("winston: skipping already imported nzb: %s", sourceNZB)
		return nil
	}

	if preview.State == StateNeedsReview && !(preview.Confidence == ConfidenceMedium && p.cfg.AutoImportMedium) {
		log.Printf("winston: review required for %s proposed=%s reason=%s", sourceNZB, preview.ProposedPath, preview.Reason)
		if p.state != nil {
			_ = p.state.Put(sourceNZB, ImportedRecord{RelativePath: preview.ProposedPath, Status: "review", State: preview.State, Confidence: preview.Confidence, Metadata: preview.Metadata, Preview: preview})
		}
		return nil
	}

	if p.alt == nil || p.alt.baseURL == "" {
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
	if p.plex != nil {
		_ = p.plex.RefreshPath(item.TargetPath)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.cfg.SleepBetweenImports):
		return nil
	}
}
