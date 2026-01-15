GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

TAGS=null

cli:
	go build -tags=$(TAGS) -mod $(GOMOD) -ldflags="-s -w" -o bin/embeddingsdb-client cmd/client/main.go
	go build -tags=$(TAGS) -mod $(GOMOD) -ldflags="-s -w" -o bin/embeddingsdb-server cmd/server/main.go

# https://developers.google.com/protocol-buffers/docs/reference/go-generated
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/org_sfomuseum_embeddingsdb_service.proto
