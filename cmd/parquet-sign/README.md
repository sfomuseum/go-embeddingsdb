# parquet-sign

For example

```
$> ./bin/parquet-sign \
	-key-uri file:///path/to/private/gpg.key \
	-password-uri 'constant://?val=s33kret' \
	/usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet
```