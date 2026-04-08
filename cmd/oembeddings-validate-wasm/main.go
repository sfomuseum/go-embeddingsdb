//go:build wasmjs
package main

import (
	"log"
	"syscall/js"

	"github.com/sfomuseum/go-embeddingsdb/oembeddings"
)

func main() {

	validate_func := oembeddings.ValidateFunc()
	defer validate_func.Release()

	js.Global().Set("oembeddings_validate", validate_func)

	c := make(chan struct{}, 0)

	log.Println("wof_format function initialized")
	<-c
}
