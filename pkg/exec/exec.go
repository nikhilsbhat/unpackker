// Package unexec is a thin wrapper around os/exec to support Unpackker.
package unexec

import (
	"io"
	"os/exec"
)

// ExecCmd implements methods required to invoke shell.
type ExecCmd struct {
	Command string
	Args    []string
	Writer  io.Writer
}

// GetCmdExec gets the constructed shell command ready to be executed..
func (e *ExecCmd) GetCmdExec() (*exec.Cmd, error) {
	cmd, err := e.getExecutable()
	if err != nil {
		return nil, err
	}
	shellCmd := &exec.Cmd{
		Path:   cmd,
		Args:   e.Args,
		Stdout: e.Writer,
		Stderr: e.Writer,
	}
	return shellCmd, nil
}

func (e *ExecCmd) getExecutable() (string, error) {
	goExecPath, err := exec.LookPath(e.Command)

	if err != nil {
		return "", err
	}
	return goExecPath, nil
}
