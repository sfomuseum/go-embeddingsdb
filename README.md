# go-embeddingsdb

An opinionated Go package for storing, indexing and querying vector embeddings.

## Motivation

## Caveats

This package and the tools it exports still occupy the in-between state of being general purpose tools and specific to the immediate needs of SFO Museum. That means it may not do what you need it to out of the box. If it doesn't we're certainly open to entertaining changes.

## Concepts

### Records

```
// Record defines a struct containing properties associated with individual records stored in an embeddings database.
type Record struct {
	// Provider is the name (or context) of the provider responsible for DepictionId.
	Provider string `json:"provider"`
	// DepictionId is the unique identifier for the depiction for which embeddings have been generated.
	DepictionId string `json:"depiction_id"`
	// SubjectId is the unique identifier associated with the record that DepictionId depicts.
	SubjectId string `json:"subject_id"`
	// Model is the label for the model used to generate embeddings for DepictionId.
	Model string `json:"model"`
	// Embeddings are the embeddings generated for DepictionId using Model.
	Embeddings []float32 `json:"embeddings"`
	// Created is the Unix timestamp when Embeddings were generated.
	Created int64 `json:"created"`
	// Attributes is an arbitrary map of key-value properties associated with the embeddings.
	Attributes map[string]string `json:"attributes"`
}
```

### Databases

```
// Database defines an interface for adding and querying vector embeddings of [embeddingsdb.Record] records.
type Database interface {
	// Add adds a [embeddingsdb.Record] instance to the underlying database implementation.
	AddRecord(context.Context, *embeddingsdb.Record) error
	// Return the EmbeddingsDB instance record matching 'provider', 'depiction_id' and 'model'.
	GetRecord(context.Context, string, string, string) (*embeddingsdb.Record, error)
	// Find similar records for a given model and record instance.
	SimilarRecords(context.Context, *embeddingsdb.SimilarRequest) ([]*embeddingsdb.SimilarResult, error)
	// Export the contents of the database. Where and how a database is exported are left as details for specific implementations.
	Export(context.Context, string) error
	// Return the Unix timestamp of the last update to the Database instance.
	LastUpdate(context.Context) (int64, error)
	// Return the URI string used to instantiate the Database instance.
	URI() string
	// Close performs and terminating functions required by the database.
	Close(context.Context) error
}
```

### Servers

```
// Server defines an interface for a network-based interface for interacting with an embeddings database.
type Server interface {
	// ListenAndServe starts a new server and listens for requests.
	ListenAndServe(context.Context) error
}
```

### Clients

```
// Clientdefines an interface for clients to interact with an embeddings database.
type Clientinterface {
	// Add a new record to an embeddings database.
	AddRecord(context.Context, *embeddingsdb.Record) error
	// Retrieve a specific record from an embeddings database.
	GetRecord(context.Context, string, string, string) (*embeddingsdb.Record, error)
	// Retrieve records with similar embeddings from an embeddings database.
	SimilarRecords(context.Context, *embeddingsdb.SimilarRequest) ([]*embeddingsdb.SimilarResult, error)
	// Retrieve records with similar embeddings, for a specific record, from an embeddings database.
	SimilarRecordsById(context.Context, string, string, string) ([]*embeddingsdb.SimilarResult, error)
}
```

## Databases

### duckdb://

Manage embeddings use the [DuckDB](https://duckdb.org/) database and the [VSS](https://duckdb.org/docs/stable/core_extensions/vss) extension.

```
duckdb://{PATH}?{QUERY_PARAMETERS}
```

Where `{PATH}` is an optional value mapped to the location of an existing DuckDB database. If present this database will be used to instantiate the database. Depending on the size of the database this can take a noticeable amount of time.

## Servers

### grcp://

Create a gRPC-based server for managing embeddings-related operations. Servers are created using a URI-based syntax as follows:

```
grpc://{HOST}:{ADDRESS}?{QUERY_PARAMETERS}
```

Valid parameters are:

| Key | Value | Required | Notes |
| --- | --- | --- | --- |
| database-uri | string | yes | A registered `sfomuseum/go-embeddingsdb/database.Database` URI for the underlying database implementation to use. |
| token-uri | string | no | A registered `gocloud.dev/runtimevar` URI used to stored a shared authentication to require with client requests. |
| tls-certificate | string | no | The path to a valid TLS certificate to use for encrypted connections. |
| tls-key | string | no | The path to a valid TLS key file to use for encrypted connections. |

```
grpc://localhost:8080?database-uri=database-uri=duckdb:///usr/local/data/embeddings&token-uri=constant%3A%2F%2F%3Fval%3Ds33kret
```

## Clients

### grpc://

Create a gRPC-based client for managing embeddings-related operations. Clients are created using a URI-based syntax as follows:

```
grpc://{HOST}:{ADDRESS}?{QUERY_PARAMETERS}
```

Valid parameters are:

| Key | Value | Required | Notes |
| --- | --- | --- | --- |
| token-uri | string | no | A registered `gocloud.dev/runtimevar` URI used to stored a shared authentication to require with client requests. |
| tls-certificate | string | no | The path to a valid TLS certificate to use for encrypted connections. |
| tls-key | string | no | The path to a valid TLS key file to use for encrypted connections. |
| tls-ca-certificate | string | no | The path to a custom TLS authority certificate to use for encrypted connections. |
| tls-insecure | bool | no | Skip TLS verification steps. Use with caution. |

For example:

```
grpc://localhost:8080?token-uri=constant%3A%2F%2F%3Fval%3Ds33kret
```

### database://

Create a client with a direct database connection for managing embeddings-related operations. Clients are created using a URI-based syntax as follows:

```
database://?{QUERY_PARAMETERS}
```

Valid parameters are:

| Key | Value | Required | Notes |
| --- | --- | --- | --- |
| database-uri | string | yes | A registered `sfomuseum/go-embeddingsdb/database.Database` URI for the underlying database implementation to use. |

For example:

```
database://?database-uri=duckdb:///usr/local/data/embeddings
```

## Tools

The easiest way to build the included tools is to run the handy `cli` Makefile target. For example:

```
$> make cli
go build -tags=duckdb -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-client cmd/client/main.go
go build -tags=duckdb -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-server cmd/server/main.go
```

### Build tags

This package uses build tags to enable support for various features. The default set of tags is `duckdb` but you can override those defaults by passing in a custom `TAGS` variable when calling the Makefile targets.

#### duckdb

The `duckdb` tag adds support for the [DuckDB](https://duckdb.org/) database as an embeddings database.

It also uses the [duckdb/duckdb-go](https://github.com/duckdb/duckdb-go) package for interacting with DuckDB in Go. Although this package bundles all its dependencies in the `vendor` folder there is one notable exception: Any of the `.a` files included in the `duckdb-go` package. That is because it add a couple hundred megabytes to the overall package size. As such you will need to run `go run tidy && go mod vendor` before compiling tools. It's not ideal but it is what it is.

Note: If you need to build a binary tool with support for DuckDB for MacOS _and_ that been signed and notarized you will need to build a customized `libduckdb_bundle.a` from source. See below [for details](#).

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

### embeddingsdb-client

```
$> ./bin/embeddingsdb-client -h
Command-line tool for interacting with a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	./bin/embeddingsdb-client [command] [options]

Valid commands are:
* record [options]
* similar-by-id [options]
```

#### embeddingsdb-client record

```
$> ./bin/embeddingsdb-client record -h
Command-line tool for retrieving a record from a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	record [options]

Valid options are:
  -client-uri string
    	A valid sfomuseum/go-mobileclip.EmbeddingsClient URI. (default "grpc://localhost:8080")
  -depiction-id string
    	The unique depiction ID associated with the record to retrieve.
  -model string
    	The name of the model associated with the record to retrieve. (default "apple/mobileclip_s0")
  -provider string
    	The name of the provider associated with the record to retrieve.
  -verbose
    	Enable vebose (debug) logging.
```

#### embeddingsdb-client query-by-id

```
$> ./bin/embeddingsdb-client similar-by-id -h
Command-line tool for retrieving records similar to the embeddings for a specific record stored in a gRPC EmbeddingsDB "service". Results are written as a JSON-encoded string to STDOUT.
Usage:
	similar-by-id [options]

Valid options are:
  -client-uri string
    	A valid sfomuseum/go-mobileclip.EmbeddingsClient URI. (default "grpc://localhost:8080")
  -depiction-id string
    	The unique depiction ID associated with the record to retrieve to establish embeddings to compare.
  -model string
    	The name of the model associated with the record to retrieve to establish embeddings to compare. (default "apple/mobileclip_s0")
  -provider string
    	The name of the provider associated with the record to retrieve to establish embeddings to compare.
  -similar-provider string
    	The name of the provider to limit similar record queries to. If empty then all the records for the model chosen will be queried.
  -verbose
    	Enable vebose (debug) logging.
```	

## DuckDB

### Statically linked extensions (MacOS)

If you want to build a `emeddingsdb-server` binary (or any other tool that uses this package as a library) for MacOS with support for DuckDB _and_ that has been signed and notarized you will need to compile a custom `libduckdb_bundle.a` library with both the JSON and VSS extensions statically linked. Then you will need to use specify that custom library when building the `emeddingsdb-server` binary. This is because the default behaviour for DuckDB is to load (and cache) extensions on the fly and those extensions will have been signed by someone other than the "team" (you) that notarized the `emeddingsdb-server` binary.

_Note: The following instructions will work if you don't care about notarizing the `emeddingsdb-server` binary but still want local, statically-linked extensions that don't require a network connection to use._

After a fair amount of trial and error this is what I managed to get working. It _should_ work for you but you know how these things end up changing when you're not looking.

First install both `duckdb` and `vcpkg` from source:

```
$> git clone https://github.com/duckdb/duckdb.git /usr/local/src/duckdb
$> git clone https://github.com/microsoft/vcpkg.git /usr/local/src/vcpkg

$> cd /usr/local/src/duckdb
```

Now copy the `vss.cmake` in to the root directory:

```
$> cp .github/config/extensions/vss.cmake ./vss_config.cmake
```

Now edit it to remove the `DONT_LINK` instruction. For example:

```
duckdb_extension_load(vss
        LOAD_TESTS
        GIT_URL https://github.com/duckdb/duckdb-vss
        GIT_TAG c8a4efe05003d8ef6eaad34f5521cf50126c9967
        TEST_DIR test/sql
        APPLY_PATCHES
    )
```

Ensure the following environment variables are set:

```
$> printenv

GEN=ninja
BUILD_VSS=1
BUILD_JSON=1
EXTENSION_CONFIGS=vss_config.cmake
VCPKG_TOOLCHAIN_PATH=/usr/local/src/vcpkg/scripts/buildsystems/vcpkg.cmake
VCPKG_ROOT=/usr/local/src/vcpkg
```

First, build the command line tool so you can verify that the VSS (and JSON) extensions are statically linked:

```
$> make

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/duckdb
 43M	/usr/local/src/duckdb/build/release/duckdb
```

Now check the installed (and loaded) extensions:

```
$> /usr/local/src/duckdb/build/release/duckdb

DuckDB v1.5.0-dev5476 (Development Version, 1c62e11b82)
Enter ".help" for usage hints.

memory D SELECT extension_name, loaded, installed, install_mode FROM duckdb_extensions() WHERE installed = true;
┌────────────────┬─────────┬───────────┬───────────────────┐
│ extension_name │ loaded  │ installed │   install_mode    │
│    varchar     │ boolean │  boolean  │      varchar      │
├────────────────┼─────────┼───────────┼───────────────────┤
│ core_functions │ true    │ true      │ STATICALLY_LINKED │
│ json           │ true    │ true      │ STATICALLY_LINKED │
│ parquet        │ true    │ true      │ STATICALLY_LINKED │
│ shell          │ true    │ true      │ STATICALLY_LINKED │
│ vss            │ true    │ true      │ STATICALLY_LINKED │
└────────────────┴─────────┴───────────┴───────────────────┘
```

Build DuckDB again as a library:

```
$> make bundle-library

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/libduckdb_bundle.a
 79M	/usr/local/src/duckdb/build/release/libduckdb_bundle.a
```

Apply extra MacOS hoop-jumping, appending the `generated_extension_loader.cpp.o` file to the `libduckdb_bundle.a` file::

```
$> find /usr/local/src/duckdb/build/release -name "generated_extension_loader.cpp.o"
/usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o

$> ar rcs /usr/local/src/duckdb/build/release/libduckdb_bundle.a /usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o
```

Finally rebuild the `embeddingsdb-server` with the customized DuckDB library:

```
$> make server-bundle
CGO_ENABLED=1 CPPFLAGS="-DDUCKDB_STATIC_BUILD" CGO_LDFLAGS="-L/usr/local/src/duckdb/build/release -lduckdb_bundle -lc++" \
	go build -tags=duckdb,duckdb_use_static_lib -mod vendor -ldflags="-s -w" \
	-o bin/embeddingsdb-server cmd/server/main.go
```

## See also
