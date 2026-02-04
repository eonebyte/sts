package main

import (
	"log"
	"log/slog"
	"os"
	"sts/web_service/internal/app"
	"sts/web_service/internal/shared/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal meload config: %v", err)
	}

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	application, err := app.NewApp(cfg, logger)

	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Server Stop: %v", err)
	}
}
