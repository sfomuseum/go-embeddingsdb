## embeddingsdb-inspector

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

![](../../docs/images/embeddingsdb-list-3.png)

You can filter the list view by model and by provider (the source of embeddings). As you can see the list view needs some loving to collapse similar depictions with multiple models in a single view. Soon, I hope.

Individual record pages look like this:

![](../../docs/images/embeddingsdb-record-3.png)

By default record pages will show similar records for a single model across all providers. Both of these facets may be updated. The left hand panel (the record being viewed) will remain fixed but the right hand panel (containing similar records) will scroll.

If enabled (with the `-enable-search` flag) there is also an endpoint where you can query for similar results by text or by uploading an image of your choosing, generate embeddings on the fly for that image and then use those data to search for similar images in the `embeddingsdb` database. For example:

![](../../docs/images/embeddingsdb-search-3.png)

As with the record view, the left hand panel (the image that was uploaded) will remain fixed but the right hand panel (containing similar records) will scroll. You can also search for images by text:

![](../../docs/images/embeddingsdb-search-text-3.png)

### Note and caveats

#### embeddingsdb-inspector is a client

Conceptually, the `embeddingsdb-inspector` is a _client_ (as described above) of an `embeddingsdb` database instance. That means one of two things:

1. You will need to have an `embeddingsdb` server instance running somewhere which will broker communications with the underlying database; for example the `grpc://localhost:8081` URI above.
2. You will need to specify a `database://` client URI appropriate to your setup; for example, to interact directly with a local DuckDB database your client URI would be something like `database://?database-uri=duckdb:///usr/local/data/embeddings`.

#### search 

In order for the search functionality to work you will need to instantiate an instance of the [sfomuseum/go-embeddings](https://github.com/sfomuseum/go-embeddings) `Client` interface. The `go-embeddingsdb` package only supports storing, indexing and querying vector embeddings. It does handle _creating_ them. This is handled by the `go-embeddings` package which supports [a number of different implementations](https://github.com/sfomuseum/go-embeddings?tab=readme-ov-file#implementations) for generating vector embeddings.

#### importing records

The `embeddingsdb-inspector` does not handle _importing_ records in to an `embeddingsdb` database. This is handled by separate processes like the [`parquet-import` tool](../parquet-import/README.md).

