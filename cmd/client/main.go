package main

import (
	"context"
	"encoding/json"
	_ "flag"
	"log"
	"os"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	args := os.Args

	if len(args) < 2 {
		log.Fatal("SAD")
	}

	cmd := args[1]

	switch cmd {
	case "-h":
		log.Fatal("help")
	case "record":
		record(args[2:])
	default:
		log.Fatal("Nope")
	}
}

func record(args []string) {

	var client_uri string
	var depiction_id int64
	var model string

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
	fs.Int64Var(&depiction_id, "depiction_id", 0, "...")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the MobileCLIP model to use to derive embeddings. Valid options are: s0, s1, s2, blt")

	fs.Parse(args)

	ctx := context.Background()

	cl, err := embeddingsdb.NewEmbeddingsDBClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	rsp, err := cl.GetRecord(ctx, depiction_id, model)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}
