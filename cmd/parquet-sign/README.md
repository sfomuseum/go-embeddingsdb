## parquet-sign

Generate a corresponding Parquet "signature" file with detached PGP/GPG signatures for embeddingsdb.Record records in one or more Parquet files.

```
$> ./bin/parquet-sign -h
Generate a corresponding Parquet "signature" file with detached PGP/GPG signatures for embeddingsdb.Record records in one or more Parquet files.
Usage:
	./bin/parquet-sign [options] parquet_file(N) parquet_file(N)
Valid options are:
  -key-uri string
    	A registered gocloud.dev/runtimevar URI which is expected to resolve to an ASCII‑armored key block.
  -output string
    	The path where Parquet-encoded data should be written. If "-" then data will be written to STDOUT.
  -password-uri string
    	A registered gocloud.dev/runtimevar URI which is expected to resolve to the key's password. This is only necessary if the key is locked and, as such, may be left empty.
  -verbose
    	Enable vebose (debug) logging.
  -verify
    	Verify signature before recording. (default true)
```

For example

```
$> ./bin/parquet-sign \
	-key-uri file:///path/to/private/gpg.key \
	-password-uri 'constant://?val=wubwub' \
	-output sfomuseum-collection-1152-siglip2-naflex-220604230-sig.parquet
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet
```