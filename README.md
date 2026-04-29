# go-embeddingsdb

An opinionated Go package for storing, indexing and querying vector embeddings.

## Motivation

There are many vector databases or databases with support for managing vector embeddings. This is not another one. This is, instead, an opinionated Go package for storing, indexing and querying vector embeddings independent of the underlying database using a common interface. Currently efforts are focused on the DuckDB-backed database (using the VSS extension) and a gRPC client/server implementation. The code, as writen, should make it easy enough to support other implementations but those have not been written yet.

This package and the tools it exports still occupy the in-between state of being general purpose and specific to the immediate needs of SFO Museum. That means it may not do what you need it to out of the box. If it doesn't we're certainly open to entertaining changes.

For background, please consult the [Similar object images derived using the MobileCLIP computer-vision models](https://millsfield.sfomuseum.org/blog/2026/01/09/similar/) blog post.

## Documentation

At this time `godoc` documentation is incomplete.

## Concepts

### Records

Records contain individual embeddings values and related metadata. While not specific to image embeddings they are what most of the work modeling records reflects.

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
	// Attributes is an arbitrary map of key-value properties associated with the embeddings. Record attributes
	// are encouraged to include the required [OEmbeddings] fields but this is not a requirement.
	Attributes map[string]string `json:"attributes"`
}
```

### Databases

A database is a system for managing (storing, indexing and querying) embeddings. This package aims to be agnostic to the underlying database system focusing instead on a common interface for use.

```
// Database defines an interface for adding and querying vector embeddings of [embeddingsdb.Record] records.
type Database interface {
	// Add adds a [embeddingsdb.Record] instance to the underlying database implementation.
	AddRecord(context.Context, *embeddingsdb.Record) error
	// The number of batched records currently waiting to be added.
	BatchedRecordsCount(context.Context) (int, error)
	// Add the pending batched records
	AddBatchedRecords(context.Context) error	
	// Return the EmbeddingsDB instance record matching 'provider', 'depiction_id' and 'model'.
	GetRecord(context.Context, *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error)
	// Remove a record from an EmbeddingsDB instance.
	RemoveRecord(context.Context, *embeddingsdb.RemoveRecordRequest) error
	// ListRecords returns a pagination list of records stored in the database.
	ListRecords(context.Context, pagination.Options, ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error)
	// IterateRecords returns an [iter.Seq2[*embeddingsdb.Record, error]] for each record stored in the database.
	IterateRecords(context.Context) iter.Seq2[*embeddingsdb.Record, error]
	// Find similar records for a given model and record instance.
	SimilarRecords(context.Context, *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error)
	// Export the contents of the database. Where and how a database is exported are left as details for specific implementations.
	Export(context.Context, string) error
	// Return the Unix timestamp of the last update to the Database instance.
	LastUpdate(context.Context) (int64, error)
	// Return the URI string used to instantiate the Database instance.
	URI() string
	// Return the unique list of models, for zero (all) or more providers, across all the embeddings.
	Models(context.Context, ...string) ([]string, error)
	// Return the unique list of providers across all the embeddings.
	Providers(context.Context) ([]string, error)
	// Close performs and terminating functions required by the database.
	Close(context.Context) error
}
```

### Servers

A server is a network-based service for managing (storing, indexing and querying) embeddings. This package aims to be agnostic to the underlying server semantics focusing instead on a common interface for use.

```
// Server defines an interface for a network-based interface for interacting with an embeddings database.
type Server interface {
	// ListenAndServe starts a new server and listens for requests.
	ListenAndServe(context.Context) error
}
```

### Clients

A client communicates with a server for managing (storing, indexing and querying) embeddings. This package aims to be agnostic to the underlying client semantics focusing instead on a common interface for use.

```
// Client defines an interface for clients to interact with an embeddings database.
type Client interface {
	// Add a new record to an embeddings database.
	AddRecord(context.Context, *embeddingsdb.Record) error
	// Retrieve a specific record from an embeddings database.
	GetRecord(context.Context, *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error)
	// Remove a record from an EmbeddingsDB instance.
	RemoveRecord(context.Context, *embeddingsdb.RemoveRecordRequest) error
	// ListRecords returns a pagination list of records stored in the database.
	ListRecords(context.Context, pagination.Options, ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error)
	// Retrieve records with similar embeddings from an embeddings database.
	SimilarRecords(context.Context, *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error)
	// Retrieve records with similar embeddings, for a specific record, from an embeddings database.
	SimilarRecordsById(context.Context, *embeddingsdb.SimilarRecordsByIdRequest) ([]*embeddingsdb.SimilarRecord, error)
	// Return the unique list of models, for zero (all) or more providers, across all the embeddings.
	Models(context.Context, ...string) ([]string, error)
	// Return the unique list of providers across all the embeddings.
	Providers(context.Context) ([]string, error)
}
```

### OEmbeddings

_Note: "OEmbeddings" should still be considered work in progress and subject to review and suggestions._

OEmbeddings defines a model for the _least_ amount of metadata to be associated with a vector embedding record in order to allow a preview of the content used to create the embeddings and to display provenance for that content with links back to the subject depicted in the content on a provider's website.

As the name suggests it is modeled in spirit after the [OEmbed specification](https://oembed.com/) which descibes itself as "a format for allowing an embedded representation of a URL on third party sites.". The `Oembeddings` structure (propeties) MAY be present in the free-form "attributes" dictionary of a `Record` instance but is not required.

```
type OEmbeddings struct {
	// The type of material used to create the vector embeddings. Expected to be "image" or "text".
	Type string `json:"type"`
	// The preview content for the vector embeddings. If `Type` is "text" then this is expected to be a string. If `Type` is "image" this is expected to be a string confirming to the JSON Schema "uri" type.
	Preview string `json:"preview"`
	// A web page (or resource) for the depiction used to create the vector embeddings.
	DepictionURL string `json:"depiction_url,omitempty"`
	// A web page (or resource) for the subject of the depiction used to create the vector embeddings.
	SubjectURL string `json:"subject_url"`
	// The title of the subject of the depiction.
	SubjectTitle string `json:"subject_title"`
	// The creditline or attribution for the subject of the depiction.
	SubjectCreditline string `json:"subject_creditline"`
	// The name of the provider (holder) of the subject being depicted.
	ProviderName string `json:"provider_name"`
	// The primary web page for the provider (holder) of the subject being depicted.
	ProviderURL string `json:"provider_url"`
}
```

#### JSON Schema

There is a [JSON Schema document](oembeddings/oembeddings.json) for validating an "attributes" dictionary to ensure that it contains the required fields for an `OEmbeddings` data structure.

#### WebAssembly

There is also an `oembeddings_validate` WebAssembly (WASM) binary for use with JavaScript. For example:

```
const input = document.querySelector("#input");
const feedback = document.querySelector("#feedback");

const oe = input.value;

oembeddings_validate(oe).then((rsp) => {
	feedback.innerText = "Document validates as OEmbeddings";
}).catch((err) => {
	console.error("Validation failed");
	feedback.innerText = "Validation failed: " + err;
});
```

The WASM binary needs to be built manually using the `make wasmjs` Makefile target. See the [oembeddings/www](oembeddings/www) folder for details.

## Databases

Here's the "tl;dr" so far:

The DuckDB implementation is generally faster than the SQLite but requires that all your data be stored in memory. That data is periodically exported to disk in order that it may be re-imported without indexing all the data from scratch but it takes a noticeable amount of time to import that data at start up time. The SQLite implementation while slower stores (and reads) all its data from disk.

The Bleve implementation is also fast, has a fast start-up time, doesn't require loading all the data in to memory, doesn't use an unmanageable amount of disk space but remains a chore to set up because of the dependency on `libfaiss` (see details below). It's also unclear to me whether it is possible to create a single, bundled executable of the Bleve implementation because of the `libfaiss` depedency.

Database documentation has been moved in to [database/README.md](database/README.md]

## Servers

Server documentation has been moved in to [server/README.md](server/README.md]

## Clients

Client documentation has been moved in to [client/README.md](client/README.md]

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

Tools documentation has been moved in to [cmd/README.md](cmd/README.md]

## DuckDB

### Statically linked extensions (MacOS)

If you want to build a `emeddingsdb-server` binary (or any other tool that uses this package as a library) for MacOS with support for DuckDB _and_ that has been signed and notarized you will need to compile a custom `libduckdb_bundle.a` library with both the JSON and VSS extensions statically linked. Then you will need to use specify that custom library when building the `emeddingsdb-server` binary. This is because the default behaviour for DuckDB is to load (and cache) extensions on the fly and those extensions will have been signed by someone other than the "team" (you) that notarized the `emeddingsdb-server` binary.

After a fair amount of trial and error this is what I managed to get working. It _should_ work for you but you know how these things end up changing when you're not looking.

_Note: There are [known problems with this process using recent releases of DuckDB](https://github.com/duckdb/duckdb-spatial/issues/794). I am trying to figure them out._

First install both `duckdb` and `vcpkg` from source:

```
$> git clone https://github.com/duckdb/duckdb.git /usr/local/src/duckdb
$> git clone https://github.com/microsoft/vcpkg.git /usr/local/src/vcpkg

$> cd /usr/local/src/duckdb
```

Now copy the `vss.cmake` config file in to the root directory:

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

Note the use of the `BUILD_JSON` environment variable. This will bundle the JSON extension which is necessary to use the VSS extension.

Now build the command line tool so you can verify that the VSS (and JSON) extensions are statically linked:

```
$> make

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/duckdb
 43M	/usr/local/src/duckdb/build/release/duckdb
```

Once built, check the installed (and loaded) extensions:

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

Assuming that the `vss` extension is installed and loaded build DuckDB again as a library:

```
$> make bundle-library

... stuff happens

$> du -h /usr/local/src/duckdb/build/release/libduckdb_bundle.a
 79M	/usr/local/src/duckdb/build/release/libduckdb_bundle.a
```

Apply additional MacOS hoop-jumping, appending the `generated_extension_loader.cpp.o` file to the `libduckdb_bundle.a` file::

```
$> find /usr/local/src/duckdb/build/release -name "generated_extension_loader.cpp.o"
/usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o

$> ar rcs /usr/local/src/duckdb/build/release/libduckdb_bundle.a /usr/local/src/duckdb/build/release/extension/CMakeFiles/duckdb_generated_extension_loader.dir/__/codegen/src/generated_extension_loader.cpp.o
```

Finally rebuild the `embeddingsdb-server` with the customized DuckDB library using the handy `server-bundle` Makefile target (in this repo):

```
$> cd /usr/local/src/go-embeddingsdb
$> mkdir work
$> cp /usr/local/src/duckdb/build/release/libduckdb_bundle.a ./work/

$> make server-bundle
CGO_ENABLED=1 CPPFLAGS="-DDUCKDB_STATIC_BUILD" CGO_LDFLAGS="-L./work -lduckdb_bundle -lc++" \
	go build -tags=duckdb,duckdb_use_static_lib -mod vendor -ldflags="-s -w" \
	-o bin/embeddingsdb-server cmd/server/main.go
```

_Note: You don't have to copy `libduckdb_bundle.a` in to a local `work` folder but this way you don't have remember where it is or what happened to it the next time you clean up your `/usr/local/src` directory. The `work` directory is explicitly excluded from Git checkins in this repository._

## See also

* https://github.com/sfomuseum/go-embeddings
* https://github.com/sfomuseum/swift-mobileclip