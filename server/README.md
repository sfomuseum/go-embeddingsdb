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

For example:

```
grpc://localhost:8080?database-uri=database-uri=duckdb:///usr/local/data/embeddings&token-uri=constant%3A%2F%2F%3Fval%3Ds33kret
```

