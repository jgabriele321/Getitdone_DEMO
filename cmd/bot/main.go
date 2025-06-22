package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/giovannigabriele/go-todo-bot/internal/config"
	"github.com/giovannigabriele/go-todo-bot/internal/health"
	"github.com/giovannigabriele/go-todo-bot/internal/llm"
	"github.com/giovannigabriele/go-todo-bot/internal/queue"
	"github.com/giovannigabriele/go-todo-bot/internal/sheets"
	"github.com/giovannigabriele/go-todo-bot/internal/telegram"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msg("Starting TODO Bot")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create LLM client
	llmClient := llm.NewClient(cfg.OpenRouterAPIKey)

	// Create Google Sheets client
	sheetsClient := sheets.NewClient(cfg.GoogleScriptURL)

	// Create queue manager
	queueManager, err := queue.NewManager(cfg.DatabasePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create queue manager")
	}
	defer queueManager.Close()

	// Create batch-capable Telegram handler only if token is provided and valid
	var handler *telegram.BatchHandler
	if cfg.TelegramToken != "" {
		// Validate the token before attempting to create the handler
		if err := telegram.ValidateTelegramToken(cfg.TelegramToken); err != nil {
			log.Warn().Err(err).Msg("Telegram token validation failed - running in web-only mode")
		} else {
			handler, err = telegram.NewBatchHandler(cfg.TelegramToken, llmClient, sheetsClient, queueManager)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to create Telegram handler - running in web-only mode")
			} else {
				log.Info().Msg("Telegram handler initialized successfully")
			}
		}
	} else {
		log.Info().Msg("Telegram token not provided - running in web-only mode")
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start health check server
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", health.Handler())

		port := os.Getenv("PORT")
		if port == "" {
			port = "10000"
		}

		server := &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		}

		log.Info().Str("port", port).Msg("Starting health check server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Health check server error")
		}
	}()

	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	// Start the bot only if handler exists
	if handler != nil {
		log.Info().Msg("Starting TODO bot...")
		if err := handler.Start(ctx); err != nil && err != context.Canceled {
			log.Fatal().Err(err).Msg("Bot error")
		}
	} else {
		log.Info().Msg("Running in web-only mode - Telegram bot disabled")
		// Keep the service running for web requests
		<-ctx.Done()
	}

	log.Info().Msg("Bot shutdown complete")
}

// setupLogging configures the logger
func setupLogging() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339

	// Use console writer for development
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	})

	// Set log level based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("Logging configured")
}
