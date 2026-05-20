package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var verbose bool

	fs := flagset.NewFlagSet("import")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "...\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	for _, path := range fs.Args() {

		r, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		defer r.Close()

		info, _ := os.Stat(path)

		f, err := parquet.OpenFile(r, info.Size())

		if err != nil {
			log.Fatal(err)
		}

		meta := f.Metadata()

		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode(meta)

		if err != nil {
			log.Fatal(err)
		}
	}
}
