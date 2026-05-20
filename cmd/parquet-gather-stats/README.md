## parquet-gather-stats

Gather embeddingsdb statistics from one or more Parquet files and write to STDOUT as JSON-encoded data..

```
$> ./bin/parquet-gather-stats -h
Gather embeddingsdb statistics from one or more Parquet files and write to STDOUT as JSON-encoded data..
Usage:
	./bin/parquet-gather-stats [options] parquet_file(N) parquet_file(N)
Valid options are:
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-gather-stats /usr/local/data/embeddings/nga/20260502-nga-opendata-1152-siglip.parquet | jq
{
  "models": [
    "google/siglip2-so400m-patch16-naflex"
  ],
  "providers": [
    "nga"
  ],
  "model_providers": {
    "google/siglip2-so400m-patch16-naflex": [
      "nga"
    ]
  },
  "model_dimensions": {
    "google/siglip2-so400m-patch16-naflex": 1152
  },
  "provider_models": {
    "nga": [
      "google/siglip2-so400m-patch16-naflex"
    ]
  }
}
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

