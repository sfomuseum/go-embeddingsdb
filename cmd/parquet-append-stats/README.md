## parquet-append-stats

Append go-embeddingsdb statistics to one or more Parquet files.

```
$> ./bin/parquet-append-stats -h
Append go-embeddingsdb statistics to one or more Parquet files.
Usage:
	./bin/parquet-append-stats [options] parquet_file(N) parquet_file(N)
Valid options are:
  -output string
    	The path where Parquet-encoded data should be written. If "-" then data will be written to STDOUT. (default "-")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-append-stats -output test2.parquet test.parquet
```

And then:

```
$> ./bin/parquet-metadata -key-value ./test2.parquet | jq
{
  "embeddingsdb:model:google/siglip2-so400m-patch14-384:providers": "sfomuseum-data-media-collection",
  "embeddingsdb:models": "google/siglip2-so400m-patch14-384",
  "embeddingsdb:provider:sfomuseum-data-media-collection:models": "google/siglip2-so400m-patch14-384",
  "embeddingsdb:providers": "sfomuseum-data-media-collection"
}
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

