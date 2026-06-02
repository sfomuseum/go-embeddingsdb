## parquet-verify

Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.

```
$> ./bin/parquet-verify -h
Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.
Usage:
	./bin/parquet-verify [options] parquet_file(N) parquet_file(N)
Valid options are:
  -public-key-uri string
    	A registered gocloud.dev/runtimevar URI that resolves to the GPG public key (armored) use to verify signatures.
  -signature value
    	One or more Parquet files containing signature data (for example, as produced by the parquet-sign tool).
  -verbose
    	Enable vebose (debug) logging.
  -workers int
    	The maximum number of concurrent worker to verify records with. (default 10)
```

For example:

```
$> ./bin/parquet-verify \
	-public-key-uri file:///usr/local/sfomuseum/go-embeddingsdb/work/sfomuseum-debug.pub \
	-signature sig-test.parquet \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet
	
2026/06/01 17:12:04 INFO Processing count=29519 valid=29509 invalid=0 missing=0 errors=0
2026/06/01 17:13:04 INFO Processing count=58646 valid=58636 invalid=0 missing=0 errors=0
2026/06/01 17:13:38 INFO Completed count=74700 valid=74700 invalid=0 missing=0 errors=0
```