package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := LoadConfig()
	app := NewApp(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("winston: starting, mode=%s sleep_between_imports=%s", cfg.DefaultMode, cfg.SleepBetweenImports)
	if err := app.Run(ctx); err != nil {
		log.Fatalf("winston: %v", err)
	}
}
