## parquet-export

Export embeddingsdb records as Parquet-encoded data.

```
$> ./bin/parquet-export -h
Export embeddingsdb records as Parquet-encoded data.
Usage:
	./bin/parquet-export [options]Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -output string
    	The path where Parquet-encoded data should be written. If "-" then data will be written to STDOUT. (default "-")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-export -output export.parquet -verbose
2026/03/30 12:15:31 DEBUG Verbose logging enabled
2026/03/30 12:15:31 DEBUG Allow insecure connections
2026/03/30 12:15:31 DEBUG Start pagination "start page"=1 "end page"=-1 "per page"=1000
2026/03/30 12:15:31 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=1 "total page count"=0
2026/03/30 12:15:34 DEBUG Assign total pages "start page"=1 "end page"=-1 "per page"=1000 pages=0
2026/03/30 12:15:34 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=2 "total page count"=236
2026/03/30 12:15:34 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=3 "total page count"=236
2026/03/30 12:15:34 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=4 "total page count"=236
...time passes
```

And then:

```
$> duckdb
DuckDB v1.4.2 (Andium) 68d7555f68
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
D SELECT COUNT(depiction_id) FROM read_parquet('export.parquet');
┌─────────────────────┐
│ count(depiction_id) │
│        int64        │
├─────────────────────┤
│       235200        │
└─────────────────────┘
```
