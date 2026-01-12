package main

import (
	"context"
	"flag"
	"os"
	"encoding/json"
	"log"
	
	"github.com/sfomuseum/go-embeddingsdb"	
)

func main() {

	var client_uri string
	var depiction_id int64
	var model string

	flag.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
	flag.Int64Var(&depiction_id, "depiction_id", 0, "...")
	flag.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the MobileCLIP model to use to derive embeddings. Valid options are: s0, s1, s2, blt")

	flag.Parse()

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
