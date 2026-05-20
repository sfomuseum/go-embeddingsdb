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
$> ./bin/parquet-gather-stats  ./test2.parquet | jq
{
  "models": [
    "google/siglip2-so400m-patch14-384"
  ],
  "providers": [
    "sfomuseum-data-media-collection"
  ],
  "model_providers": {
    "google/siglip2-so400m-patch14-384": [
      "sfomuseum-data-media-collection"
    ]
  },
  "provider_models": {
    "sfomuseum-data-media-collection": [
      "google/siglip2-so400m-patch14-384"
    ]
  }
}
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

