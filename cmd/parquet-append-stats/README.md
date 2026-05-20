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
$> ./bin/parquet-append-stats -output nga.parquet /usr/local/data/embeddings/nga/20260413-nga-opendata.parquet
```

And then:

```
$> ./bin/parquet-metadata -key-value ./nga.parquet | jq
{
  "embeddingsdb:model:apple/mobileclip_s0:dimensions": "512",
  "embeddingsdb:model:apple/mobileclip_s0:providers": "nga",
  "embeddingsdb:model:apple/mobileclip_s1:dimensions": "512",
  "embeddingsdb:model:apple/mobileclip_s1:providers": "nga",
  "embeddingsdb:model:apple/mobileclip_s2:dimensions": "512",
  "embeddingsdb:model:apple/mobileclip_s2:providers": "nga",
  "embeddingsdb:models": "apple/mobileclip_s2;apple/mobileclip_s1;apple/mobileclip_s0",
  "embeddingsdb:provider:nga:models": "apple/mobileclip_s2;apple/mobileclip_s1;apple/mobileclip_s0",
  "embeddingsdb:providers": "nga"
}
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

