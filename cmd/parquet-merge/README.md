## parquet-merge

Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.

```
$> ./bin/parquet-merge -h
Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.
Usage:
	./bin/parquet-merge [options] parquet_file(N) parquet_file(N)
Valid options are:
  -output string
    	The path where Parquet-encoded data should be written. If "-" then data will be written to STDOUT. (default "-")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-merge \
	-verbose \
	-output merged.parquet \
	../go-embeddings-harvest/sfomuseum-collection-siglip2-naflex.parquet \
	../go-embeddings-harvest/sfomuseum-ig-siglip2-naflex.parquet
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

