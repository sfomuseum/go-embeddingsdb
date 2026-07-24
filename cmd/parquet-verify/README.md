## parquet-verify

Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.

```
$> ./bin/parquet-verify -h
Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.
Usage:
	./bin/parquet-verify [options] parquet_file(N) parquet_file(N)
Valid options are:
<<<<<<< HEAD
  -signatures value
    	One or more Parquet files containing signature data (for example, as produced by the parquet-sign tool).
  -verbose
    	Enable vebose (debug) logging.
  -verifier-uri string
    	A valid sfomuseum/go-embeddingsdb/signatures.Verifier URI.
=======
  -public-key-uri string
    	A registered gocloud.dev/runtimevar URI that resolves to the GPG public key (armored) use to verify signatures.
  -signature value
    	One or more Parquet files containing signature data (for example, as produced by the parquet-sign tool).
  -verbose
    	Enable vebose (debug) logging.
>>>>>>> 86fe2b437fce7ac70eb0a9ff5187be241bee257b
  -workers int
    	The maximum number of concurrent worker to verify records with. (default 10)
```

<<<<<<< HEAD
_For details on the form that that the `-verifier-uri` flag should take consult the [signatures/README.md](../../signatures/README.md) documentation._

=======
>>>>>>> 86fe2b437fce7ac70eb0a9ff5187be241bee257b
For example:

```
$> ./bin/parquet-verify \
<<<<<<< HEAD
	-signatures sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.parquet \
	-verifier-uri 'tls://?certificate-uri=sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.pub' \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet

2026/06/03 15:00:05 INFO Completed count=74700 valid=74700 invalid=0 missing=0 errors=0
=======
	-public-key-uri file:///usr/local/sfomuseum/go-embeddingsdb/work/sfomuseum-debug.pub \
	-signature sig-test.parquet \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet
	
2026/06/01 17:12:04 INFO Processing count=29519 valid=29509 invalid=0 missing=0 errors=0
2026/06/01 17:13:04 INFO Processing count=58646 valid=58636 invalid=0 missing=0 errors=0
2026/06/01 17:13:38 INFO Completed count=74700 valid=74700 invalid=0 missing=0 errors=0
>>>>>>> 86fe2b437fce7ac70eb0a9ff5187be241bee257b
```