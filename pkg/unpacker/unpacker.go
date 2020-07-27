package unpacker

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/terragen/decode"
	"github.com/nikhilsbhat/unpackker/pkg/backend"
	unexec "github.com/nikhilsbhat/unpackker/pkg/exec"
	"github.com/nikhilsbhat/unpackker/pkg/helper"
)

//UnPackkerInput holds the required fields to unpack the asset.
type UnPackkerInput struct {
	// Name of the packed asset.
	Name string `json:"name" yaml:"name"`
	// StubPath refers to path where the client stub is placed.
	StubPath string `json:"stubpath" yaml:"stubpath"`
	// TargetPath refers to path where the asset has to be unpacked.
	TargetPath string `json:"assetpath" yaml:"assetpath"`
	// CleanStub make sure that the client stub is removed by Unpacker after successful unpacking of asset.
	CleanStub bool `json:"cleanstub" yaml:"cleanstub"`
	// Writer to be assigned so that Unpacker can logs its outputs and errors.
	// AssetBackend for the asset generated.
	AssetBackend *backend.Store
	Writer       io.Writer
	version      string
	cmd          *unexec.ExecCmd
}

// NewConfig retunrns new config of UnPackkerInput.
func NewConfig() *UnPackkerInput {
	return &UnPackkerInput{}
}

// Unpacker unpacks the asset onto specified path.
func (i *UnPackkerInput) Unpacker() error {
	if err := i.validate(); err != nil {
		return err
	}

	i.cmd = i.getRootCmd()

	if err := i.fetchAsset(); err != nil {
		return err
	}

	if err := i.unpackAsset(); err != nil {
		return err
	}

	i.cleanClientStub()
	return nil
}

func (i *UnPackkerInput) validate() error {
	i.generateDefaults()
	i.TargetPath = i.getAssetPath(i.TargetPath)
	if len(i.StubPath) != 0 {
		i.StubPath = i.getAssetPath(i.StubPath)
	}

	// This is responsible fetching the artifact from remote backed to path specified
	if err := i.fetchAssetFromRemote(); err != nil {
		return err
	}

	stub, err := os.Stat(i.StubPath)
	if err != nil {
		return fmt.Errorf("Unable to find the Stub at specified location")
	}
	if stub.IsDir() {
		return fmt.Errorf("Stub/Asset cannot be directory")
	}

	if err := i.initBackend(); err != nil {
		return err
	}
	return nil
}

func (i *UnPackkerInput) generateDefaults() {
	if i.Writer == nil {
		i.Writer = os.Stdout
	}
	if len(i.TargetPath) == 0 {
		i.TargetPath = "."
	}
}

func (i *UnPackkerInput) getAssetPath(assetpath string) string {
	if assetpath == "." {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		return dir
	}
	abstPath, err := filepath.Abs(path.Dir(assetpath))
	if err != nil {
		fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
		os.Exit(1)
	}
	return filepath.Join(abstPath, helper.SplitBasePath(assetpath))
}

func (i *UnPackkerInput) initBackend() error {
	if len(i.StubPath) != 0 {
		i.AssetBackend.TargetPath = i.StubPath
		return nil
	}
	if err := i.AssetBackend.Backend(); err != nil {
		return err
	}
	if len(i.AssetBackend.TargetPath) == 0 {
		i.AssetBackend.TargetPath = i.StubPath
	}
	return nil
}

// Even fetchAssetFrom unpacks asset, additional task it does is to fecth the artifact from remote path.
func (i *UnPackkerInput) fetchAssetFromRemote() error {
	return nil
}

func (i *UnPackkerInput) fetchAsset() error {
	if err := i.AssetBackend.FetchAsset(); err != nil {
		return err
	}
	return nil
}

func (i *UnPackkerInput) unpackAsset() error {
	if i.cmd != nil {
		args := append(i.cmd.Args, "generate", "--path", i.TargetPath)
		newCmd := i.cmd
		newCmd.Args = args
		cmd, err := newCmd.GetCmdExec()
		if err != nil {
			return err
		}
		output, err := cmd.Output()
		if len(output) != 0 {
			return fmt.Errorf(string(output))
		}
		if err != nil {
			return fmt.Errorf("Oops..! an error occured while unpacking %v", err)
		}
		return nil
	}
	return fmt.Errorf("Oops..! an error occured while unpacking asset")
}

func (i *UnPackkerInput) fetchAssetVersion() error {
	if i.cmd != nil {
		args := append(i.cmd.Args, "version", "-s")
		newCmd := i.cmd
		newCmd.Args = args
		cmd, err := newCmd.GetCmdExec()
		if err != nil {
			return err
		}
		output, err := cmd.Output()
		if len(output) != 0 {
			return fmt.Errorf(string(output))
		}
		if err != nil {
			return err
		}
		i.version = string(output)
		return nil
	}
	return fmt.Errorf("Oops..! an error occured while fetching version of asset")
}

func (i *UnPackkerInput) getRootCmd() *unexec.ExecCmd {
	exe := new(unexec.ExecCmd)
	exe.Command = i.AssetBackend.TargetPath
	exe.Args = []string{"generate"}
	//exe.Writer = i.Writer
	return exe
}

func (i *UnPackkerInput) cleanClientStub() {
	if i.CleanStub {
		fmt.Println(ui.Info("Cleaning the client stub is initiated as unpack is successful\n"))
		err := os.RemoveAll(i.AssetBackend.TargetPath)
		if err != nil {
			fmt.Println(ui.Error(fmt.Sprintf("Oops..! an error occured while cleaning the mess at %s, you have to clear it before next run", i.AssetBackend.TargetPath)))
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		fmt.Println(ui.Info("All files and folders created by Unpaccker in the process of packing asset was cleared\n"))
	}
}
