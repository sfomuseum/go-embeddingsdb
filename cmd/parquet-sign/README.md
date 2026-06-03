## parquet-sign

Generate a corresponding Parquet "signature" file with detached PGP/GPG signatures for embeddingsdb.Record records in one or more Parquet files.

```
$> ./bin/parquet-sign -h
Generate a corresponding Parquet "signature" file with PGP/GPG or TLS detached signatures for embeddingsdb.Record records in one or more Parquet files.
Usage:
	./bin/parquet-sign [options] parquet_file(N) parquet_file(N)
Valid options are:
  -embed-public-key
    	If true the public key (certicate) used to verify signatures will be written to the signature Parquet file in the "embeddingsdb:signatures:public_key" metadata key.
  -signer-uri string
    	A valid sfomuseum/go-embeddingsdb/signatures.Signer URI.
  -target-bucket-uri string
    	The URI where signature files and public keys will be written. One of the following: A valid gocloud.dev/blob.Bucket URI; The path to a folder on the local filesystem; "cwd://" which will cause files to be written to the current directory. (default "cwd://")
  -verbose
    	Enable vebose (debug) logging.
  -verify
    	Verify signature before recording. (default true)
```

For example:

```
$> ./bin/parquet-sign \
	-verbose \
	-signer-uri 'tls://?certificate-uri=file:///usr/local/sfomuseum/tls-certificate.pub&key-uri=file:///usr/local/sfomuseum/tls-certificate.key' \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet
	
2026/06/03 14:41:00 DEBUG Verbose logging enabled
2026/06/03 14:42:00 INFO Processing count=52815 signed=52814 verified=52814 completed=52814 errors=0
2026/06/03 14:42:25 INFO Completed count=74700 signed=74700 verified=74700 completed=74700 errors=0
```

This will produce two files:

1. A Parquet file containing hashed `embeddingsdb.Record` entries and a "detached" signature.
2. The public key (or certificate) which can be used to verify those signatures.

For example:

```
$> ls -al sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.*
-rw-r--r--  1 asc  staff  250945 Jun  3 14:42 sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.parquet
-rw-r--r--  1 asc  staff    2589 Jun  3 14:42 sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.pub
```

The naming conventions for these files is to strip the ".parquet" extension from the input file and then append "-signatures.parquet" or "-signatures.pub" to the signature and public key files, respectively.

To use those signatures and the public key to verify records you would do something like this:

```
$> ./bin/parquet-verify \
	-signatures sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.parquet \
	-verifier-uri 'tls://?certificate-uri=sfomuseum-collection-1152-siglip2-naflex-22060423-signatures.pub' \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet

2026/06/03 15:00:05 INFO Completed count=74700 valid=74700 invalid=0 missing=0 errors=0
```

_See the [cmd/parquet-verify](../parquet-verify) documentation for details.