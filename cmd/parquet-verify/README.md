## parquet-verify

Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.

```
$> ./bin/parquet-verify -h
Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.
Usage:
	./bin/parquet-verify [options] parquet_file(N) parquet_file(N)
Valid options are:
  -signatures value
    	One or more Parquet files containing signature data (for example, as produced by the parquet-sign tool).
  -verbose
    	Enable vebose (debug) logging.
  -verifier-uri string
    	A valid sfomuseum/go-embeddingsdb/signatures.Verifier URI.
  -workers int
    	The maximum number of concurrent worker to verify records with. (default 10)
```

_For details on the form that that the `-verifier-uri` flag should take consult the [signatures/README.md](../../signatures/README.md) documentation._

For example:

```
$> ./bin/parquet-verify \
	-signatures sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.parquet \
	-verifier-uri 'tls://?certificate-uri=sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.pub' \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet

2026/06/03 15:00:05 INFO Completed count=74700 valid=74700 invalid=0 missing=0 errors=0
```