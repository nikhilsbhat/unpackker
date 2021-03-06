// Package packer handles the process of packaging asset with the support of supporting library.
package packer

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-bindata/go-bindata/v3"
	"github.com/imdario/mergo"
	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/terragen/decode"
	"github.com/nikhilsbhat/unpackker/pkg/backend"
	gen "github.com/nikhilsbhat/unpackker/pkg/gen"
	"github.com/nikhilsbhat/unpackker/pkg/helper"
	"github.com/spf13/cobra"
)

//PackkerInput holds the required fields to pack the asset.
type PackkerInput struct {
	// The name of the asset client stub.
	Name string `json:"name" yaml:"name"`
	// TempPath would be used to carryout all the operation of Unpackker defaults to PWD.
	TempPath string `json:"tempath" yaml:"tempath"`
	// Path to asset which has to be packed.
	AssetPath string `json:"assetpath" yaml:"assetpath" env:"UNPACKKER_ASSET_PATH"`
	// AssetMetaData is set of metadata that has to be applied to the asset.
	// This value can be used while unpacking or for any future requirements.
	AssetMetaData map[string]string `json:"assetmetadata" yaml:"assetmetadata"`
	// Path defines where the packed asset has to be placed.
	Path string `json:"path" yaml:"path"`
	// IgnoreFiles are regexes of the files that should be avided.
	IgnoreFiles []string `json:"ignore" yaml:"ignore"`
	// Environment in which the asset is packed.
	Environment string `json:"environment" yaml:"environment" env:"UNPACKKER_ENVIRONMENT"`
	// AssetVersion refers to version of asset which has to eb packed.
	AssetVersion string `json:"assetversion" yaml:"assetversion" env:"UNPACKKER_ASSET_VERSION"`
	// Backend for the asset generated.
	Backend *backend.Store `json:"backend" yaml:"backend"`
	// ConfigPath refers to file path where the config file lies, defaults to PWD.
	ConfigPath string `json:"configpath" yaml:"configpath" env:"UNPACKKER_CONFIG_PATH"`
	// CleanLocalCache clears the local cache creted under PackkerInput.Path if enabled,
	// this will be effective only if backend is type 'fs'.
	CleanLocalCache bool `json:"cleancache" yaml:"cleancache" env:"UNPACKKER_CLEAN_LOCALCACHE"`
	// targetPath refers to path where the packed asset has to be placed.
	targetPath     string
	filesToIgnore  []*regexp.Regexp
	clinetStubPath string
	gen.GenInput
	// writer   io.Writer
}

// NewConfig retunrns new config of PackkerInput.
func NewConfig() *PackkerInput {
	return &PackkerInput{}
}

// Packer packs the asset which would be understood by unpacker
func (i *PackkerInput) Packer(cmd *cobra.Command, args []string) {
	configFromFile, err := i.LoadConfig()
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err) + "\n"))
		os.Exit(1)
	}

	if configFromFile != nil {
		if err := mergo.Merge(configFromFile, i, mergo.WithOverride); err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
	}

	cfg, envcfg, err := getConfigFromEnvWithValidate()
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		fmt.Println(ui.Warn("Dropping env variables as we ran into problem while fetching"))
	}
	if envcfg {
		fmt.Println(ui.Warn("Dropping env variables as no corresponding values set"))
	} else {
		if err := mergo.Merge(configFromFile, cfg, mergo.WithOverride); err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
	}

	if configFromFile == nil {
		configFromFile = i
	}

	if err := configFromFile.validate(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		configFromFile.cleanMess()
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
		configFromFile.cleanMess()
		os.Exit(1)
	}

	configFromFile.clinetStubPath = filepath.Join(configFromFile.TempPath, clientStub)
	if err := configFromFile.buildAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		configFromFile.cleanMess()
		os.Exit(1)
	}

	// Setup clientstub to make it ready for packaging
	fmt.Println(ui.Info("Unpackker is in the process of packing asset\n"))
	if err := configFromFile.setupAssetDir(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		configFromFile.cleanMess()
		os.Exit(1)
	}

	fmt.Println(ui.Info("Prerequisites for Asset packing is completed successfully\n"))

	if err := configFromFile.packAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		configFromFile.cleanMess()
		os.Exit(1)
	}

	fmt.Println(ui.Warn(configFromFile.nameForTemp()), ui.Info(" was packed successfully\n"))

	fmt.Println(ui.Info("Storing packed asset onto the specified backend\n"))
	if err := configFromFile.storeAsset(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		configFromFile.cleanMess()
		os.Exit(1)
	}
	fmt.Println(ui.Info("Asset was stored successfully, it should be available in the backed configured\n"))

	configFromFile.cleanMess()
}

func (i *PackkerInput) validate() error {
	i.generateDefaults()
	i.Path = i.getPath()
	i.TempPath = i.getTempPath() + "_temp"

	if err := i.getFilesToIgnore(); err != nil {
		return err
	}

	if i.tempPathExists() {
		return fmt.Errorf("looks like Unpackker was exited abruptly which left behind few traces at %s\nIt will be cleared now", i.TempPath)
	}

	if err := i.createTempPath(); err != nil {
		return fmt.Errorf(decode.GetStringOfMessage(err))
	}

	if !(i.assetExists()) {
		return fmt.Errorf("could not find the asset here %s, either user does not permission or wrong path specified", i.AssetPath)
	}

	targetPath, err := filepath.Abs(fmt.Sprintf("%s/%s", path.Dir(i.TempPath), i.nameForTemp()))
	if err != nil {
		return err
	}
	i.targetPath = targetPath

	if err := i.initBackend(); err != nil {
		return err
	}

	return nil
}

func (i *PackkerInput) getFilesToIgnore() error {
	if len(i.IgnoreFiles) != 0 {
		patterns := make([]*regexp.Regexp, 0)
		i.IgnoreFiles = append(i.IgnoreFiles, helper.SplitBasePath(i.Path), helper.SplitBasePath(i.TempPath))
		fmt.Println(ui.Info(fmt.Sprintf("Files that would be exempted are: %v \n", i.IgnoreFiles)))
		for _, pattern := range i.IgnoreFiles {
			patterns = append(patterns, regexp.MustCompile(pattern))
		}
		i.filesToIgnore = patterns
		return nil
	}
	return fmt.Errorf("nothing to ignore, or defaults are not set right")
}

func (i *PackkerInput) initBackend() error {
	if err := i.Backend.InitBackend(); err != nil {
		return err
	}
	if i.Backend == nil {
		i.Backend = backend.New()
	}
	if len(i.Backend.Path) == 0 {
		i.Backend.Path = i.targetPath
	}
	if len(i.Backend.Name) == 0 {
		i.Backend.Name = i.Name
	}
	if len(i.Backend.MetaData) == 0 {
		i.Backend.MetaData = i.AssetMetaData
	}
	return nil
}

func (i *PackkerInput) storeAsset() error {
	i.Backend.Folder = filepath.ToSlash(filepath.Join(i.Backend.Folder, i.Backend.Name))
	i.Backend.Name = i.nameForTemp()
	if err := i.Backend.StoreAsset(); err != nil {
		return err
	}
	return nil
}

func (i *PackkerInput) generateDefaults() {
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
	if len(i.AssetVersion) == 0 {
		i.AssetVersion = "1.0"
	}
	if i.Backend == nil {
		newbackend := backend.New()
		newbackend.Cloud = "fs"
		i.Backend = newbackend
	}

	i.IgnoreFiles = append(i.IgnoreFiles, helper.SplitBasePath(i.ConfigPath))
}

func (i *PackkerInput) setupAssetDir() error {
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

func (i *PackkerInput) packAsset() error {
	goBuild := exec.Command("go", "build", "-o", i.Backend.Path, "-ldflags", "-s -w")
	goBuild.Dir = i.clinetStubPath

	if err := goBuild.Run(); err != nil {
		return err
	}
	return nil
}

func (i *PackkerInput) getPath() string {
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

func (i *PackkerInput) getTempPath() string {
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

func (i *PackkerInput) nameForTemp() string {
	res := strings.ReplaceAll(i.AssetVersion, ".", "_")
	return fmt.Sprintf("%s_%s", i.Name, res)
}

func (i *PackkerInput) createTempPath() error {
	err := os.MkdirAll(i.TempPath, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (i *PackkerInput) assetExists() bool {
	if _, direrr := os.Stat(i.AssetPath); os.IsNotExist(direrr) {
		return false
	}
	return true
}

func (i *PackkerInput) tempPathExists() bool {
	if _, direrr := os.Stat(i.TempPath); os.IsNotExist(direrr) {
		return false
	}
	return true
}

func (i *PackkerInput) buildAsset() error {
	cfg := bindata.NewConfig()

	cfg.Ignore = i.filesToIgnore
	cfg.Prefix = helper.SplitBasePath(i.AssetPath, "base")
	cfg.Output = filepath.Join(i.clinetStubPath, i.Name, helper.SplitBasePath(i.clinetStubPath)+".go")
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

func (i *PackkerInput) cleanMess() {
	if !i.CleanLocalCache {
		fmt.Println(ui.Warn("Cleaning traces was skipped, as cleancache is disabled. Make sure to clean it manually before next run\n"))
		os.Exit(1)
	}

	fmt.Println(ui.Info("Cleaning the mess created while packing the asset\n"))
	err := os.RemoveAll(i.TempPath)
	if err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("oops..! an error occurred while cleaning the traces at %s: %v\n, ", i.TempPath, err)))
		fmt.Println(ui.Error("it should be to cleared manually before next run"))
		os.Exit(1)
	}

	if err := i.cleanCache(); err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
	}
	fmt.Println(ui.Info("All files and folders created by Unpaccker in the process of packing asset was cleared successfully\n"))
}

func (i *PackkerInput) cleanCache() error {
	if i.Backend.Cloud != "fs" {
		err := os.RemoveAll(i.Path)
		if err != nil {
			return err
		}
	}
	return nil
}
