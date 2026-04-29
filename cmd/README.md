## Tools

The easiest way to build the included tools is to run the handy `cli` Makefile target. For example:

```
$> make cli
go build -tags=sqlite -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-client cmd/client/main.go
go build -tags=sqlite -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-server cmd/server/main.go
go build -tags=sqlite -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-inspector cmd/inspector/main.go
go build -tags=sqlite -mod vendor -ldflags="-s -w" -o bin/parquet-export cmd/parquet-export/main.go
go build -tags=sqlite -mod vendor -ldflags="-s -w" -o bin/parquet-import cmd/parquet-import/main.go
go build -tags=sqlite -mod readonly -ldflags="-s -w" -o bin/parquet-merge cmd/parquet-merge/main.go
```

### DuckDB

DuckDB is a dependency regardless of build tags (described below).

This package uses the [duckdb/duckdb-go](https://github.com/duckdb/duckdb-go) package for interacting with DuckDB in Go. Although this package bundles all its dependencies in the `vendor` folder there is one notable exception: Any of the `.a` files included in the `duckdb-go` package. That is because it add a couple hundred megabytes to the overall package size. As such you will need to run `go run tidy && go mod vendor` before compiling tools. It's not ideal but it is what it is.

Note: If you need to build a binary tool with support for DuckDB for MacOS _and_ that been signed and notarized you will need to build a customized `libduckdb_bundle.a` from source. See below [for details](#statically-linked-extensions-macos).

### Build tags

Build tags are used to enable support for various features. The default set of tags are `sqlite` but you can override those defaults by passing in a custom `TAGS` variable when calling the Makefile targets.

#### bleve

The `bleve` tag adds support for [Bleve](https://blevesearch.com/) document store as an embeddings database. Note that the `vectors` tags is also necessary.

#### sqlite

The `sqlite` tag adds support for the [SQLite](https://sqlite.org/) database as an embeddings database. This uses the [sqlite-vec](https://alexgarcia.xyz/sqlite-vec/) extension for vector embeddings support.

_Note: As of this writing only the Go-language [CGO bindings](https://github.com/asg017/sqlite-vec-go-bindings?tab=readme-ov-file#cgo-bindings) are supported. Support for "pure Go" bindings will be added in future releases._

#### vectors

The `vectors` tag is necessary to compile `libfaiss` code when building Bleve document store support. This is a compliement to the `bleve` tag.

### embeddingsdb-server

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

### embeddingsdb-client

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

#### embeddingsdb-client record

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

#### embeddingsdb-client remove

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

#### embeddingsdb-client similar-by-id

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

#### embeddingsdb-client list

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

#### embeddingsdb-client models

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

#### embeddingsdb-client providers

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

### embeddingsdb-inspector

A minimalist web-interface for inspecting documents stored in a `embeddingsdb-server` instance.

```
$> ./bin/embeddingsdb-inspector -h
A minimalist web-interface for inspecting documents stored in a `embeddingsdb-server` instance.
Usage:
	./bin/embeddingsdb-inspector [options]
Valid options are:
  -client-uri string
    	A validsfomuseum/go-embeddingsdb/client.Client URI. (default "grpc://localhost:8080")
  -embeddings-client-uri string
    	A registered go-embeddings.Client URI. This is required if the -enable-search flag is true.
  -enable-search
    	Enable search functionality.
  -max-results int
    	The maximum number of similar results to return. (default 20)
  -max-upload-size int
    	The maximum size (in bytes) for uploads. (default 10485760)
  -server-uri string
    	A registered aaronland/go-http/v4/server.Server URI. (default "http://localhost:8080")
  -uri-prefix string
    	An optional prefix (location) to serve the application from.
  -verbose
    	Enable verbose (debug) logging.
```

For example:

```
$> make inspector
go run -tags=sqlite -mod vendor \
		cmd/inspector/main.go \
		-verbose \
		-client-uri 'grpc://localhost:8081' \
		-enable-search \
		-embeddings-client-uri 'mobileclip://?client-uri=grpc://localhost:8080' \
		-server-uri http://localhost:8082
2026/03/30 12:42:01 DEBUG Verbose logging enabled
2026/03/30 12:42:01 DEBUG Allow insecure connections
2026/03/30 12:42:01 INFO Listen for requests address=http://localhost:8082
```

Opening your web browser to `http://localhost:8082` you would see something like this (depending on the records you've indexed in the `embeddingsdb` databae):

![](../docs/images/embeddingsdb-list-3.png)

You can filter the list view by model and by provider (the source of embeddings). As you can see the list view needs some loving to collapse similar depictions with multiple models in a single view. Soon, I hope.

Individual record pages look like this:

![](../docs/images/embeddingsdb-record-3.png)

By default record pages will show similar records for a single model across all providers. Both of these facets may be updated. The left hand panel (the record being viewed) will remain fixed but the right hand panel (containing similar records) will scroll.

If enabled (with the `-enable-upload` flag) there is also an endpoint where you can upload an image of your choosing, generate embeddings on the fly for that image and then use those data to search for similar images in the `embeddingsdb` database. For example:

![](../docs/images/embeddingsdb-search-3.png)

As with the record view, the left hand panel (the image that was uploaded) will remain fixed but the right hand panel (containing similar records) will scroll. You can also search for images by text:

![](../docs/images/embeddingsdb-search-text-3.png)

#### Note and caveats

##### embeddingsdb-inspector is a client

Conceptually, the `embeddingsdb-inspector` is a _client_ (as described above) of an `embeddingsdb` database instance. That means one of two things:

1. You will need to have an `embeddingsdb` server instance running somewhere which will broker communications with the underlying database; for example the `grpc://localhost:8081` URI above.
2. You will need to specify a `database://` client URI appropriate to your setup; for example, to interact directly with a local DuckDB database your client URI would be something like `database://?database-uri=duckdb:///usr/local/data/embeddings`.

##### search 

In order for the search functionality to work you will need to instantiate an instance of the [sfomuseum/go-embeddings](https://github.com/sfomuseum/go-embeddings) `Client` interface. The `go-embeddingsdb` package only supports storing, indexing and querying vector embeddings. It does handle _creating_ them. This is handled by the `go-embeddings` package which supports [a number of different implementations](https://github.com/sfomuseum/go-embeddings?tab=readme-ov-file#implementations) for generating vector embeddings.

##### importing records

The `embeddingsdb-inspector` does not handle _importing_ records in to an `embeddingsdb` database. This is handled by separate processes like the `parquet-import` tool described below.

### parquet-import

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

### parquet-export

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

### parquet-merge

Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.

```
$> ./bin/parquet-merge -h
Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.
Usage:
	./bin/parquet-merge [options] parquet_file(N) parquet_file(N)
Valid options are:
  -output string
    	The path where Parquet-encoded data should be written. If "-" then data will be written to STDOUT. (default "-")
  -verbose
    	Enable vebose (debug) logging.
```

For example:

```
$> ./bin/parquet-merge \
	-verbose \
	-output merged.parquet \
	../go-embeddings-harvest/sfomuseum-collection-siglip2-naflex.parquet \
	../go-embeddings-harvest/sfomuseum-ig-siglip2-naflex.parquet
```

If an input URI (to merge) starts with `http(s)://` then that file will be read over the wire using DuckDB's `read_parquet` functionality.

