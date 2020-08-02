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
	unpackConfig.TargetPath = "testing/test_path"
	// unpackConfig.StubPath = "testing/test_path/asset_name_0_1_0"

	backend := backend.New()
	// Type is not required field, if not specified it uses type 'fs' by default.
	backend.Cloud = "gcp"
	backend.Bucket = "path/to/bucket"
	backend.Name = "asset_name_0_1_0"
	backend.CredentialPath = "path/to/credentail.json" // credentails.json incase of gcp
	// CredentialType is not required field, if not specified it sets to default.
	backend.CredentialType = "file"
	backend.TargetPath = unpackConfig.TargetPath

	unpackConfig.AssetBackend = backend

	if err := unpackConfig.Unpacker(); err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Println("Success")
	}
}
