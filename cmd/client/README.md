## embeddingsdb-client

Command-line tool for interacting with a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.

```
$> ./bin/embeddingsdb-client -h
Command-line tool for interacting with a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	./bin/embeddingsdb-client [command] [options]

Valid commands are:
* record [options]
* remove [options]
* similar-by-id [options]
* list [options]
* models [options]
* providers [options]
```

_Note: This tool does implement all of the `Client` interface methods (notably for adding records) yet._

### embeddingsdb-client record

Command-line tool for retrieving a record from a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.

```
$> ./bin/embeddingsdb-client record -h
Command-line tool for retrieving a record from a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	record [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -depiction-id string
    	The unique depiction ID associated with the record to retrieve.
  -model string
    	The name of the model associated with the record to retrieve. (default "apple/mobileclip_s0")
  -provider string
    	The name of the provider associated with the record to retrieve.
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/embeddingsdb-client record -provider sfomuseum-data-media-collection -depiction-id 1527858087 -client-uri 'grpc://localhost:8080' | jq
{
  "provider": "sfomuseum-data-media-collection",
  "depiction_id": "1527858087",
  "subject_id": "1511924695",
  "model": "apple/mobileclip_s0",
  "embeddings": [
    -0.017242432,
    -0.021408081,
    ... and so on
```

### embeddingsdb-client remove

Command-line tool for removing a record from a gRPC EmbeddingsDB "service".

```
$> ./bin/embeddingsdb-client remove -h
Command-line tool for removing a record from a gRPC EmbeddingsDB "service".
Usage:
	record [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -depiction-id string
    	The unique depiction ID associated with the record to retrieve.
  -model string
    	The name of the model associated with the record to retrieve. (default "apple/mobileclip_s0")
  -provider string
    	The name of the provider associated with the record to retrieve.
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/embeddingsdb-client remove -client-uri grpc://localhost:8081 -depiction-id 08_michael_ross_sfom.jpg -provider sfomuseum-dotorg-exhibition -model apple/mobileclip_s2
```

### embeddingsdb-client similar-by-id

Command-line tool for retrieving records similar to the embeddings for a specific record stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.

```
$> ./bin/embeddingsdb-client similar-by-id -h
Command-line tool for retrieving records similar to the embeddings for a specific record stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	similar-by-id [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -depiction-id string
    	The unique depiction ID associated with the record to retrieve to establish embeddings to compare.
  -max-distance float
    	The maximum distance allowed when querying records. This will override defaults established by the server.
  -max-results int
    	The maximum number of results to return in a query. This will override defaults established by the server.
  -model string
    	The name of the model associated with the record to retrieve to establish embeddings to compare. (default "apple/mobileclip_s0")
  -provider string
    	The name of the provider associated with the record to retrieve to establish embeddings to compare.
  -similar-provider string
    	The name of the provider to limit similar record queries to. If empty then all the records for the model chosen will be queried.
  -verbose
    	Enable vebose (debug) logging.
```	

For example:

```
$> ./bin/embeddingsdb-client similar-by-id -provider sfomuseum-data-media-collection -depiction-id 1527858087 -client-uri 'grpc://localhost:8081' \
	| jq -r '.[]["depiction_id"]'
	
1527858091
1527858093
1880320457
1880320459
1880320639
1914676715
1914058931
1880273579
1880319239
1964039457
```

### embeddingsdb-client list

Paginated list of all the records in an embeddingsdb database emitted to STDOUT as line-separated JSON.Usage:

```
$> ./bin/embeddingsdb-client list -h
Paginated list of all the records in an embeddingsdb database emitted to STDOUT as line-separated JSON.Usage:
	list [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -end-page int
    	The maximum page number of results to emit. If -1 then this flag will be ignored and all the results (remaining after -start-page * -per-page) will be returned. (default -1)
  -per-page int
    	The number of records to include in each paginated result set. (default 10)
  -start-page int
    	The initial page of results to emit. (default 1)
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
> ./bin/embeddingsdb-client list -verbose -per-page 1000 > test.jsonl
2026/03/30 12:04:30 DEBUG Verbose logging enabled
2026/03/30 12:04:30 DEBUG Allow insecure connections
2026/03/30 12:04:30 DEBUG Start pagination "start page"=1 "end page"=-1 "per page"=1000
2026/03/30 12:04:30 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=1 "total page count"=0
2026/03/30 12:04:33 DEBUG Assign total pages "start page"=1 "end page"=-1 "per page"=1000 pages=0
2026/03/30 12:04:33 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=2 "total page count"=236
2026/03/30 12:04:33 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=3 "total page count"=236
2026/03/30 12:04:33 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=4 "total page count"=236
2026/03/30 12:04:33 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=5 "total page count"=236
... time passes, pagination happens
2026/03/30 12:05:20 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=235 "total page count"=236
2026/03/30 12:05:21 DEBUG Query records "start page"=1 "end page"=-1 "per page"=1000 page=236 "total page count"=236

$> wc -l test.jsonl
  235200 test.jsonl
```

### embeddingsdb-client models

Command-line tool for retrieving the unique list of models stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.

```
$> ./bin/embeddingsdb-client models -h
Command-line tool for retrieving the unique list of models stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	models [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -provider value
    	Zero or more providers to limit model selection by.
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/embeddingsdb-client models -client-uri 'grpc://localhost:8081' | jq
[
  "apple/mobileclip_s0",
  "apple/mobileclip_s2",
  "apple/mobileclip_s1"
]
```

### embeddingsdb-client providers

Command-line tool for retrieving the unique list of providers stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.

```
$> ./bin/embeddingsdb-client providers -h
Command-line tool for retrieving the unique list of providers stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	models [options]

Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/embeddingsdb-client providers -client-uri 'grpc://localhost:8081' | jq
[
  "sfomuseum-data-media-collection"
]
```
