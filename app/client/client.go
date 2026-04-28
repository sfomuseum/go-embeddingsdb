package client

import (
	"context"
	"log/slog"
	"os"
)

func Run(ctx context.Context) error {

	args := os.Args

	if len(args) < 2 {
		usage()
	}

	cmd := args[1]

	switch cmd {
	case "-h":
		usage()
	case "add":
		AddRecord(ctx, args[2:])
	case "record":
		GetRecord(ctx, args[2:])
	case "remove":
		RemoveRecord(ctx, args[2:])
	case "similar-by-id":
		SimilarRecordsById(ctx, args[2:])
	case "models":
		Models(ctx, args[2:])
	case "list":
		ListRecords(ctx, args[2:])
	case "providers":
		Providers(ctx, args[2:])
	default:
		slog.Warn("Unsupported command", "command", cmd)
		usage()
	}

	return nil
}
