## parquet-metadata

Print JSON-encoded metadata for a Parquet file to STDOUT

```
$> ./bin/parquet-metadata -h
Print JSON-encoded metadata for a Parquet file to STDOUT
Usage:
	./bin/parquet-metadata [options] parquet_file
Valid options are:
  -key-value
    	Only display key-value metadata.
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-metadata  ./test2.parquet | jq '.NumRows'
74600
```

or:

```
$> ./bin/parquet-metadata  -key-value ./test2.parquet | jq
{
  "embeddingsdb:model:google/siglip2-so400m-patch14-384:providers": "sfomuseum-data-media-collection",
  "embeddingsdb:models": "google/siglip2-so400m-patch14-384",
  "embeddingsdb:provider:sfomuseum-data-media-collection:models": "google/siglip2-so400m-patch14-384",
  "embeddingsdb:providers": "sfomuseum-data-media-collection"
}
```