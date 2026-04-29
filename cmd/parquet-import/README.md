## parquet-import

Import parquet-encoded embeddingsdb records from one or more files and add them to an embeddingsdb instance.

```
$> ./bin/parquet-import -h
Import parquet-encoded embeddingsdb records from one or more files and add them to an embeddingsdb instance.
Usage:
	./bin/parquet-import [options] parquet_file(N) parquet_file(N)
Valid options are:
  -client-uri string
    	A registered sfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-import -client-uri grpc://localhost:8081 -verbose ./test.parquet 
2026/03/24 11:10:11 DEBUG Verbose logging enabled
2026/03/24 11:10:11 DEBUG Allow insecure connections
2026/03/24 11:11:11 DEBUG Records imported count=9958
...and so on
```

And then:

```
$> duckdb
DuckDB v1.4.2 (Andium) 68d7555f68
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
D SELECT COUNT(depiction_id) FROM read_parquet('test.parquet');
┌─────────────────────┐
│ count(depiction_id) │
│        int64        │
├─────────────────────┤
│       216774        │
└─────────────────────┘
```
