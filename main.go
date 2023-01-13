package main

import (
	"log"

	"github.com/rodaine/protoc-gen-visibility/runner"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	log.Default().SetPrefix("[protoc-gen-visibility] ")
	log.Default().SetFlags(log.Ltime | log.LUTC | log.Lmsgprefix)

	protogen.Options{}.Run(runner.New())
}
