package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type QueueRunner struct {
	cfg  Config
	proc *ImportProcessor
}

func NewQueueRunner(cfg Config, proc *ImportProcessor) *QueueRunner {
	return &QueueRunner{cfg: cfg, proc: proc}
}

func (q *QueueRunner) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if err := q.runOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("winston: queue pass error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (q *QueueRunner) runOnce(ctx context.Context) error {
	items, err := q.listNZBs()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	for _, nzb := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := q.proc.ImportOne(ctx, nzb); err != nil {
			return err
		}
	}
	return nil
}

func (q *QueueRunner) listNZBs() ([]string, error) {
	var out []string
	err := filepath.WalkDir(q.cfg.SourceRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".nzb") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}
