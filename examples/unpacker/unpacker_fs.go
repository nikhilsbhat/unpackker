package main

import (
	"fmt"
	"os"

	"github.com/nikhilsbhat/unpackker/pkg/backend"
	unpack "github.com/nikhilsbhat/unpackker/pkg/unpacker"
)

func main() {
	unpackConfig := unpack.NewConfig()
	unpackConfig.Writer = os.Stdout
	unpackConfig.CleanStub = false
	unpackConfig.TargetPath = "testing"
	unpackConfig.StubPath = "testing/compose_1_1"

	backend := backend.New()
	// Type is not required field, if not specified it uses type 'fs' by default.
	backend.Cloud = "fs"
	//backend.TargetPath = "testing/compose_1_0"

	unpackConfig.AssetBackend = backend
	if err := unpackConfig.Unpacker(); err != nil {
		fmt.Printf("%v", err)
	} else {
		fmt.Println("Success")
	}
}
