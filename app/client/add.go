package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sfomuseum/go-embeddings"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func AddRecord(ctx context.Context, args []string) {

	var client_uri string
	var database_uris multi.MultiString
	var embeddings_client_uri string

	var action string
	var input string

	var provider string
	var depiction_id string
	var subject_id string

	var verbose bool

	fs := flagset.NewFlagSet("add")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.StringVar(&embeddings_client_uri, "embeddings-client-uri", "", "...")

	fs.Var(&database_uris, "database-uri", "...")

	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve.")
	fs.StringVar(&subject_id, "subject-id", "", "The subject ID.")

	fs.StringVar(&action, "action", "", "Valid option are: image, text")
	fs.StringVar(&input, "input", "", "If - then data is read from STDIN.")

	// Attributes

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for adding a record to a gRPC EmbeddingsDB \"service\".")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "add")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	cl_closefunc := func() {

		ctx := context.Background()
		err := cl.Close(ctx)

		if err != nil {
			slog.Error("Failed to close", "error", err)
			log.Fatalf("Failed to close client, %v", err)
		}
	}

	defer cl_closefunc()

	emb_cl, err := embeddings.NewEmbedder32(ctx, embeddings_client_uri)

	if err != nil {
		log.Fatalf("Failed to create embeddings client, %v", err)
	}

	// START OF put me in a function

	var body []byte

	switch action {
	case "text":

		switch input {
		case "-":

			b, err := io.ReadAll(os.Stdin)

			if err != nil {
				log.Fatalf("Failed to read STDIN, %v", err)
			}

			body = b
		default:

			b, err := os.ReadFile(input)

			if err != nil {
				log.Fatalf("Failed to read file, %v", err)
			}

			body = b
		}

	case "image":

		b, err := os.ReadFile(input)

		if err != nil {
			log.Fatalf("Failed to read file, %v", err)
		}

		body = b

	default:
		log.Fatalf("Invalid action")
	}

	embeddings_req := &embeddings.EmbeddingsRequest{
		Body: body,
	}

	var embeddings_rsp any
	var embeddings_err error

	switch action {
	case "text":
		embeddings_rsp, embeddings_err = emb_cl.TextEmbeddings(ctx, embeddings_req)
	case "image":
		embeddings_rsp, embeddings_err = emb_cl.ImageEmbeddings(ctx, embeddings_req)
	default:
		log.Fatal("Invalid action")
	}

	if embeddings_err != nil {
		log.Fatal(err)
	}

	// END OF put me in a function

	e := embeddings_rsp.(embeddings.EmbeddingsResponse[float32])

	db_rec := &embeddingsdb.Record{
		Provider:    provider,
		DepictionId: depiction_id,
		SubjectId:   subject_id,
		Model:       e.Model(),
		Embeddings:  e.Embeddings(),
		Created:     e.Created(),
		// Attributes:  opts.Attributes,
	}

	err = cl.AddRecord(ctx, db_rec)

	if err != nil {
		log.Fatalf("Failed to add record, %v", err)
	}

	log.Printf("Added record %s\n", db_rec.Key())
}
