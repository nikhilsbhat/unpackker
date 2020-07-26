package packer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-bindata/go-bindata/v3"
	"github.com/imdario/mergo"
	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/terragen/decode"
	"github.com/nikhilsbhat/unpackker/pkg/backend"
	gen "github.com/nikhilsbhat/unpackker/pkg/gen"
	"github.com/spf13/cobra"
)

//UnpackkerInput holds the required values to generate the templates
type UnpackkerInput struct {
	// The name of the asset client stub.
	Name string `json:"name" yaml:"name"`
	// TempPath would be used to carryout all the operation of Unpackker defaults to PWD.
	TempPath string `json:"tempath" yaml:"tempath"`
	// Path to asset which has to be packed.
	AssetPath string `json:"assetpath" yaml:"assetpath"`
	// Path defines where the packed asset has to be placed.
	Path string `json:"path" yaml:"path"`
	// Environment in which the asset is packed.
	Environment string `json:"environment" yaml:"environment"`
	// AssetVersion refers to version of asset which has to eb packed.
	AssetVersion string `json:"assetversion" yaml:"assetversion"`
	// Backend for the asset generated.
	Backend *backend.Store `json:"backend" yaml:"backend"`
	// ConfigPath refers to file path where the config file lies, defaults to PWD.
	ConfigPath string `json:"configpath" yaml:"configpath"`
	// TargetPath refers to path where the packed asset has to be placed.
	TargetPath     string
	clinetStubPath string
	gen.GenInput
	writer   io.Writer
	template string
}

// NewConfig retunrns new config of UnpackkerInput.
func NewConfig() *UnpackkerInput {
	return &UnpackkerInput{}
}

// Packer packs the asset which would be understood by unpacker
func (i *UnpackkerInput) Packer(cmd *cobra.Command, args []string) {
	configFromFile, err := i.LoadConfig()
	if err != nil {
		fmt.Println(ui.Warn(decode.GetStringOfMessage(err) + "\n"))
		fmt.Println(ui.Warn("Switching to default config"))
	}

	// bi, err := json.MarshalIndent(configFromFile, "", " ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(bi))

	if err := mergo.Merge(configFromFile, i, mergo.WithOverride); err != nil {
		os.Exit(1)
	}

	if err := configFromFile.validate(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	genin := new(gen.GenInput)
	genin.Package = configFromFile.Name
	genin.Path = configFromFile.TempPath
	genin.Environment = configFromFile.Environment
	genin.AssetVersion = configFromFile.AssetVersion

	clientStub, err := genin.Generate()
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	configFromFile.clinetStubPath = filepath.Join(configFromFile.TempPath, clientStub)
	if err := configFromFile.buildAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	// Setup clientstub to make it ready for packaging
	fmt.Println(ui.Info("Unpackker is in the process of packing asset\n"))
	if err := configFromFile.setupAssetDir(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	fmt.Println(ui.Info("Prerequisites for Asset packing is completed successfully\n"))
	if err := configFromFile.packAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}

	fmt.Println(ui.Info("Asset was packed successfully\n"))
	fmt.Println(ui.Info("Cleaning the mess created while packing the asset\n"))

	if err := configFromFile.cleanMess(); err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Oops..! an error occured while cleaning the mess at %s, you have to clear it before next run", i.TempPath)))
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}
}

func (i *UnpackkerInput) validate() error {
	i.generateDefaults()
	i.Path = i.getPath()
	i.TempPath = i.getTempPath() + "_temp"

	if i.tempPathExists() {
		return fmt.Errorf("Looks like Unpackker was exited abruptly which left behind few traces at %s\nIt has to be cleared manually for now", i.TempPath)
	}

	if err := i.createTempPath(); err != nil {
		return fmt.Errorf(decode.GetStringOfMessage(err))
	}

	if !(i.assetExists()) {
		return fmt.Errorf("Could not find the asset here %s, either user does not permission or wrong path specified", i.AssetPath)
	}
	return nil
}

func (i *UnpackkerInput) generateDefaults() {
	if len(i.Path) == 0 {
		i.Path = "."
	}
	if len(i.AssetPath) == 0 {
		i.AssetPath = "."
	}
	if len(i.Environment) == 0 {
		i.Environment = "development"
	}
	if len(i.Name) == 0 {
		i.Name = "demo"
	}
}

func (i *UnpackkerInput) setupAssetDir() error {
	goInit := exec.Command("go", "mod", "init", i.Name)
	goInit.Dir = i.clinetStubPath
	if err := goInit.Run(); err != nil {
		return err
	}

	goVnd := exec.Command("go", "mod", "vendor")
	goVnd.Dir = i.clinetStubPath
	if err := goVnd.Run(); err != nil {
		return err
	}
	return nil
}

func (i *UnpackkerInput) packAsset() error {
	buildPath, err := filepath.Abs(fmt.Sprintf("%s/%s", path.Dir(i.TempPath), i.nameForTemp()))
	if err != nil {
		return err
	}
	goBuild := exec.Command("go", "build", "-o", buildPath, "-ldflags", "-s -w")
	goBuild.Dir = i.clinetStubPath

	if err := goBuild.Run(); err != nil {
		return err
	}
	return nil
}

func (i *UnpackkerInput) getPath() string {
	if i.Path == "." {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		return dir
	}
	unpath, err := filepath.Abs(i.Path)
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}
	return unpath
}

func (i *UnpackkerInput) getTempPath() string {
	if i.Path == "." {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		return filepath.Join(dir, i.nameForTemp())
	}
	return filepath.Join(i.Path, i.nameForTemp())
}

func (i *UnpackkerInput) nameForTemp() string {
	res := strings.ReplaceAll(i.AssetVersion, ".", "_")
	return fmt.Sprintf("%s_%s", i.Name, res)
}

func (i *UnpackkerInput) createTempPath() error {
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

	cfg.Prefix = splitBasePath(i.AssetPath, "base")
	cfg.Output = filepath.Join(i.clinetStubPath, i.Name, splitBasePath(i.clinetStubPath)+".go")
	cfg.Package = i.Name
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

func (i *UnpackkerInput) cleanMess() error {
	err := os.RemoveAll(i.TempPath)
	if err != nil {
		return err
	}
	return nil
}

func splitBasePath(path string, pathType ...string) string {
	dir, filepath := filepath.Split(path)
	if len(pathType) == 0 {
		return filepath
	}
	return dir
}
