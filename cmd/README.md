## Tools

The easiest way to build the included tools is to run the handy `cli` Makefile target. For example:

```
$> make cli
go build -tags= -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-client cmd/client/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-server cmd/server/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/embeddingsdb-inspector cmd/inspector/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-export cmd/parquet-export/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-import cmd/parquet-import/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-merge cmd/parquet-merge/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-metadata cmd/parquet-metadata/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-emit cmd/parquet-emit/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-sign cmd/parquet-sign/main.go
go build -tags= -mod vendor -ldflags="-s -w" -o bin/parquet-verify cmd/parquet-verify/main.go
```

_Please be sure to read the [notes on building tools](../#build) for build tags related to specific database implementations and other "known knowns"._

### embeddingsdb-server

Start a network-based server for managing embeddings.

Detailed documentation for this tool can be found in [server/README.md](server/README.md).

### embeddingsdb-client

Command-line tool for interacting with a gRPC EmbeddingsDB server.

Detailed documentation for this tool can be found in [client/README.md](client/README.md).

### embeddingsdb-inspector

A minimalist web-interface for inspecting documents stored in a `embeddingsdb-server` instance.

Detailed documentation for this tool can be found in [inspector/README.md](inspector/README.md).

### parquet-append-stats

Append go-embeddingsdb statistics to one or more Parquet files.

Detailed documentation for this tool can be found in [parquet-append-stats/README.md](parquet-append-stats/README.md).

### parquet-emit

Emit embeddingsdb records in a Parquet as JSON-encoded data.

Detailed documentation for this tool can be found in [parquet-emit/README.md](parquet-emit/README.md).

### parquet-gather-stats

Gather embeddingsdb statistics from one or more Parquet files and write to STDOUT as JSON-encoded data..

Detailed documentation for this tool can be found in [parquet-gather-stats/README.md](parquet-gather-stats/README.md).

### parquet-import

Import parquet-encoded embeddingsdb records from one or more files and add them to an embeddingsdb instance.

Detailed documentation for this tool can be found in [parquet-import/README.md](parquet-import/README.md).

### parquet-export

Export embeddingsdb records as Parquet-encoded data.

Detailed documentation for this tool can be found in [parquet-export/README.md](parquet-export/README.md).

### parquet-merge

Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.

Detailed documentation for this tool can be found in [parquet-merge/README.md](parquet-merge/README.md).

### parquet-metadata

Print JSON-encoded metadata for a Parquet file to STDOUT

Detailed documentation for this tool can be found in [parquet-metadata/README.md](parquet-metadata/README.md).

### parquet-sign

Generate a corresponding Parquet "signature" file with detached PGP/GPG signatures for embeddingsdb.Record records in one or more Parquet files.

Detailed documentation for this tool can be found in [parquet-sign/README.md](parquet-sign/README.md).

### parquet-verify

Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.

Detailed documentation for this tool can be found in [parquet-verify/README.md](parquet-verify/README.md).
