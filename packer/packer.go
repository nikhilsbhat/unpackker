package packer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-bindata/go-bindata/v3"
	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/terragen/decode"
	gen "github.com/nikhilsbhat/unpackker/gen"
	"github.com/spf13/cobra"
)

var (
	tempPath = "."
)

//UnpackkerInput holds the required values to generate the templates
type UnpackkerInput struct {
	// The name of the asset client stub.
	Name string
	// TempPath would be used to carryout all the operation of Unpackker defaults to PWD.
	TempPath string
	// Path to asset which has to be packed.
	AssetPath string
	// AssetVersion refers to version of asset which has to eb packed.
	AssetVersion   string
	clinetStubPath string
	gen.GenInput
	writer   io.Writer
	template string
}

// Packer packs the asset which would be understood by unpacker
func (i *UnpackkerInput) Packer(cmd *cobra.Command, args []string) {

	if len(i.AssetVersion) == 0 {
		i.AssetVersion = "1.0"
	}

	i.TempPath = i.getTempPath()
	fmt.Println(i.TempPath)

	if i.tempPathExists() {
		fmt.Println(ui.Error(fmt.Sprintf("Looks like Unpackker was exited abruptly which left behind few traces at %s\nIt has to be cleared manually for now", i.TempPath)))
		os.Exit(1)
	}

	if err := i.createTempPath(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	if !(i.assetExists()) {
		fmt.Println(ui.Error(fmt.Sprintf("Could not find the asset here %s, either user does not permission or wrong path specified\n", i.AssetPath)))
		os.Exit(1)
	}

	genin := new(gen.GenInput)
	genin.Package = i.Name
	genin.Path = i.TempPath
	genin.Environment = i.Environment

	clientStub, err := genin.Generate()
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	i.clinetStubPath = clientStub
	if err := i.buildAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	// Setup the project to make it ready for developable
	fmt.Println(ui.Info("terragen is in the process of making provider creator's life easy\n"))
	if err := i.setupAssetDir(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}
	fmt.Println(ui.Info("life of provider creator is less complicated now ...!!\nStart enhancing the code from the templates generated by terragen\n"))

	if err := i.packAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}
}

func (i *UnpackkerInput) setupAssetDir() error {
	goInit := exec.Command("go", "mod", "init", i.Name)
	goInit.Dir = i.TempPath
	if err := goInit.Run(); err != nil {
		return err
	}

	goVnd := exec.Command("go", "mod", "vendor")
	goVnd.Dir = i.TempPath
	if err := goVnd.Run(); err != nil {
		return err
	}
	return nil
}

func (i *UnpackkerInput) packAsset() error {
	goBuild := exec.Command("go", "build", "-ldflags='-s -w'", i.Name)
	goBuild.Dir = i.TempPath
	if err := goBuild.Run(); err != nil {
		return err
	}
	return nil
}

func (i *UnpackkerInput) getTempPath() string {
	if i.Path == "." {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		//i.TempPath = filepath.Join(dir, i.TempPath, i.nameForTemp())
		return filepath.Join(dir, i.TempPath, i.nameForTemp())
	}
	//i.TempPath = filepath.Join(i.Path, i.TempPath, i.nameForTemp())
	return filepath.Join(i.Path, i.TempPath, i.nameForTemp())
}

func (i *UnpackkerInput) nameForTemp() string {
	res := strings.ReplaceAll(i.AssetVersion, ".", "_")
	return fmt.Sprintf("%s_%s", i.Name, res)
}
func (i *UnpackkerInput) createTempPath() error {
	fmt.Println(ui.Info(fmt.Sprintf("template %s does not exists, generating one for you\n", i.template)))
	err := os.MkdirAll(i.TempPath, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (i *UnpackkerInput) assetExists() bool {
	if _, direrr := os.Stat(i.AssetPath); os.IsNotExist(direrr) {
		return false
	}
	return true
}

func (i *UnpackkerInput) tempPathExists() bool {
	if _, direrr := os.Stat(i.TempPath); os.IsNotExist(direrr) {
		return false
	}
	return true
}

func (i *UnpackkerInput) buildAsset() error {
	cfg := bindata.NewConfig()

	cfg.Prefix = i.TempPath
	cfg.Output = fmt.Sprintf("%s/%s.go", i.TempPath, i.clinetStubPath)
	cfg.Package = i.Package
	cfg.HttpFileSystem = true

	assetPath := []string{i.AssetPath}
	cfg.Input = make([]bindata.InputConfig, len(assetPath))
	for i := range cfg.Input {
		cfg.Input[i] = bindata.InputConfig{Path: assetPath[i], Recursive: true}
	}

	if err := bindata.Translate(cfg); err != nil {
		return err
	}
	return nil
}
