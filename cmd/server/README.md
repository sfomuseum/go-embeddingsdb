## embeddingsdb-server

Start a network-based server for managing embeddings.

```
$> ./bin/embeddingsdb-server -h
Start a network-based server for managing embeddings.
Usage:
	./bin/embeddingsdb-server [options]
Valid options are:
  -database-uri string
    	An optional value which be used to replace the '{database}' placeholder, if present, in the -server-uri flag. This is expected to be a registered sfomuseum/go-embeddingsdb/database.Database URI
  -server-uri string
    	A registered sfomuseum/go-embeddingsdb/server.EmbeddingsDBServer URI. (default "grpc://localhost:8081?database-uri={database}&token-uri={token}")
  -token-uri string
    	An optional value which be used to replace the '{token}' placeholder, if present, in the -server-uri flag. This is expected to be a registered gocloud.dev/runtimevar URI that resolves to a shared authentication token.
  -verbose
    	Enable vebose (debug) logging.
```	

For example:

```
$> ./bin/embeddingsdb-server -server-uri 'grpc://localhost:8081?database-uri={database}' -database-uri 'duckdb:///usr/local/data/embeddings' -verbose
2026/01/17 06:24:58 DEBUG Verbose logging enabled
2026/01/17 06:24:58 DEBUG Set up database
2026/01/17 06:24:58 DEBUG Statically linked VSS extension installed and loaded
2026/01/17 06:24:58 DEBUG Load database from path path=/usr/local/data/embeddings
2026/01/17 06:24:58 DEBUG IMPORT DATABASE '/usr/local/data/embeddings'
2026/01/17 06:25:40 DEBUG Finished setting up database time=41.931554166s
2026/01/17 06:25:40 DEBUG Set up database export timer path=/usr/local/data/embeddings
2026/01/17 06:25:40 DEBUG Set up listener
2026/01/17 06:25:40 DEBUG Set up server
2026/01/17 06:25:40 DEBUG Allow insecure connections
2026/01/17 06:25:40 INFO Server listening address=localhost:8081
```

_Note: Did you notice the "Statically linked VSS extension installed and loaded" message in the example above? This is NOT the default behaviour (which is to install and load the `VSS` extension on the fly, downloading it from the DuckDB servers as necessary). See below [for details](#statically-linked-extensions-macos)_ 
