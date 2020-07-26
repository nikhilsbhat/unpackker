package unpacker

import (
	"fmt"
	"io"
	"os"

	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/unpackker/pkg/backend"
	unexec "github.com/nikhilsbhat/unpackker/pkg/exec"
)

//UnPackkerInput holds the required fields to unpack the asset.
type UnPackkerInput struct {
	// Name of the packed asset.
	Name string `json:"name" yaml:"name"`
	// StubPath refers to path where the client stub is placed.
	StubPath string `json:"stubpath" yaml:"stubpath"`
	// CleanStub make sure that the client stub is removed by Unpacker after successful unpacking of asset.
	CleanStub bool `json:"cleanstub" yaml:"cleanstub"`
	// Writer to be assigned so that Unpacker can logs its outputs and errors.
	Writer io.Writer
	// Backend for the asset generated.
	Backend *backend.Store `json:"backend" yaml:"backend"`
	version string
	cmd     *unexec.ExecCmd
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
	return nil
}

func (i *UnPackkerInput) validate() error {
	if i.Writer == nil {
		i.Writer = os.Stdout
	}
	return nil
}

func (i *UnPackkerInput) fetchAssetFrom() error {
	return nil
}

func (i *UnPackkerInput) unpackAsset() error {
	if i.cmd != nil {
		i.cmd.Args = append(i.cmd.Args, "--path", i.Backend.Path)
		cmd, err := i.cmd.GetCmdExec()
		if err != nil {
			return err
		}
		_, err = cmd.Output()
		if err != nil {
			fmt.Println(ui.Error("Oops..! an error occured while unpacking"))
			return err
		}
		return nil
	}
	return fmt.Errorf("Oops..! an error occured while fetching version of asset")
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
	exe.Command = "unpackker"
	exe.Args = []string{"generate"}
	return exe
}
