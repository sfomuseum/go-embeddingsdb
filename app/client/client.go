package client

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func usage() {

	fmt.Fprintf(os.Stderr, "Command-line tool for interacting with a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
	fmt.Fprintf(os.Stderr, "Usage:\n\t%s [command] [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Valid commands are:\n")
	fmt.Fprintf(os.Stderr, "* add [options]\n")
	fmt.Fprintf(os.Stderr, "* record [options]\n")
	fmt.Fprintf(os.Stderr, "* remove [options]\n")
	fmt.Fprintf(os.Stderr, "* similar-by-id [options]\n")
	fmt.Fprintf(os.Stderr, "* list [options]\n")
	fmt.Fprintf(os.Stderr, "* models [options]\n")
	fmt.Fprintf(os.Stderr, "* providers [options]\n")
	flag.PrintDefaults()

	os.Exit(1)
}

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
