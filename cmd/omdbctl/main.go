package main

import (
	"context"
	"fmt"
	"github.com/lsmoura/omdb-api/database"
	"github.com/lsmoura/omdb-api/logging"
	"github.com/lsmoura/omdb-api/omdb"
	"golang.org/x/exp/slog"
	"os"
	"os/signal"
)

func run(ctx context.Context, args []string) error {
	if len(args) == 1 {
		return fmt.Errorf("%s: missing command", args[0])
	}

	db, err := database.DB()
	if err != nil {
		return fmt.Errorf("database.DB: %w", err)
	}

	switch args[1] {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  import-all-movies")
		fmt.Println("  import-movie-links")
	case "import-all-movies":
		if err := omdb.ImportAllMovies(ctx, db); err != nil {
			return fmt.Errorf("omdb.ImportAllMovies: %w", err)
		}
	case "import-movie-links":
		if err := omdb.ImportMovieLinks(ctx, db); err != nil {
			return fmt.Errorf("omdb.ImportMovieLinks: %w", err)
		}
	default:
		return fmt.Errorf("%s: unknown command", args[1])
	}

	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout))
	ctx := logging.WithLogger(context.Background(), logger)
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(ctx, "panic", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		cancel()
	}()

	if err := run(ctx, os.Args); err != nil {
		logger.ErrorCtx(ctx, "run", "error", err)
		os.Exit(1)
	}

	logger.InfoCtx(ctx, "done")
}
