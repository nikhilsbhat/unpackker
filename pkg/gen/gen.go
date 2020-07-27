// Package gen is the core of UnPackker, where the template generation happens.
package gen

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/nikhilsbhat/neuron/cli/ui"
)

//GenInput holds the required values to generate the templates
type GenInput struct {
	// The name of the asset client stub.
	Package string `json:"package" yaml:"package"`
	// Path defines where the templates has to be generated.
	Path        string `json:"path" yaml:"path"`
	Environment string `json:"environment" yaml:"environment"`
	// AssetVersion refers to version of asset which has to eb packed.
	AssetVersion string `json:"assetversion" yaml:"assetversion"`
	// TemplateRaw consists of go-templates which are required for generation of client stub.
	TemplateRaw UnpackkerTemplate
	// AutoGenMessage will be configured by unpackker and cannot be overwritten.
	AutoGenMessage string
	// writer         io.Writer
	template string
}

// UnpackkerTemplate are the collections of go-templates which are used to generate ClientStub of UnPackker.
type UnpackkerTemplate struct {
	// CliTemp holds the template for provider
	CliTemp string
	// RootTemp holds the template for root file
	RootTemp string
	// CliMetaTemp holds the template for data
	CliMetaTemp string
	// FlagsTemp holds the template for resource
	FlagsTemp string
	// RegisterTemp holds the template for resource
	RegisterTemp string
}

var rootTemp = `{{ .AutoGenMessage }}
// Package main initializes the cli of UnPackker client stub
package main

import (
	cli "{{ .Package }}/{{ .Package }}"
)

//This function is responsible for starting the application.
func main() {
	cli.Main()
}`

var cliTemp = `{{ .AutoGenMessage }}
package {{ .Package }}

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cmd *cobra.Command
)

func init() {
	cmd = SetPacckerStubCmds()
}

// Main will take the workload of executing/starting the cli, when the command is passed to it.
func Main() {
	err := Execute(os.Args[1:])
	if err != nil {
		cm.NeuronSaysItsError(err.Error())
		os.Exit(1)
	}
}

// Execute will actually execute the cli by taking the arguments passed to cli.
func Execute(args []string) error {

	cmd.SetArgs(args)
	_, err := cmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}`

var cliMetaTemp = `{{ .AutoGenMessage }}
package {{ .Package }}

import (
	"github.com/nikhilsbhat/neuron/cli/ui"
	"os"
)

type cliMeta struct {
	*ui.NeuronUi
}

var (
	cm = &cliMeta{}
)

func init() {

	nui := ui.NeuronUi{&ui.UiWriter{os.Stdout}}
	cm = &cliMeta{&nui}

}`

var flagsTemp = `{{ .AutoGenMessage }}
package {{ .Package }}

import (
	"github.com/spf13/cobra"
)

// Registering all the flags to the command neuron itself.
func registerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&genin.assetName, "name", "n", "{{ .Package }}", "name of the asset that needs to be unpacked")
	cmd.PersistentFlags().StringVarP(&genin.path, "path", "p", ".", "path where the asset has to be unpacked")
	cmd.PersistentFlags().BoolVarP(&genin.silent, "silent", "s", false, "silence the output to get more speccific output")
}
func registerVersionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&genin.silent, "silent", "s", false, "silence the output to get more speccific output")
}
`

var registerTemp = `{{ .AutoGenMessage }}
package {{ .Package }}

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/nikhilsbhat/neuron/cli/ui"
	"github.com/nikhilsbhat/terragen/decode"
	"github.com/spf13/cobra"
)

var (
	cmds         map[string]*cobra.Command
	genin        genInput
	assetVersion = "{{ .AssetVersion }}"
	env          = "{{ .Environment }}"
)

type confcmds struct {
	commands []*cobra.Command
}

type genInput struct {
	assetName string
	assetPath string
	version   string
	path      string
	silent    bool
	env       string
}

// SetPacckerStubCmds helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetPacckerStubCmds() *cobra.Command {
	cmd := getPackkerCmd()
	return cmd
}

func getPackkerCmd() *cobra.Command {

	var packkerCmd = &cobra.Command{
		Use:   "unpackker [command]",
		Short: "Mode of asset unpacking using client stub",
		// Args:  cobra.MinimumNArgs(1),
		RunE: cm.echoPackker,
	}
	// packkerCmd.SetUsageTemplate(getUsageTemplate())

	var genCmd = &cobra.Command{
		Use:          "generate [flags]",
		Short:        "Command to generate the asset on to specified folder",
		Run:          genin.generate,
		SilenceUsage: true,
	}

	// fetching "version" will be done here.
	var versionCmd = &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of asset that is bundled along with this binary",
		RunE:  genin.versionConfig,
	}

	genCmd.Hidden = true
	versionCmd.Hidden = true
	registerFlags(genCmd)
	registerVersionFlags(versionCmd)
	packkerCmd.AddCommand(genCmd)
	packkerCmd.AddCommand(versionCmd)
	return packkerCmd
}

func (cm *cliMeta) echoPackker(cmd *cobra.Command, args []string) error {
	fmt.Println(ui.Warn("This binary is expected to invoked by unpackker library, use unpackker to get use of it"))
	return nil
}

// This function will return the custom template for usage function,
// only functions/methods inside this package can call this.

func (i *genInput) versionConfig(cmd *cobra.Command, args []string) error {
	if i.silent {
		fmt.Println(assetVersion)
		return nil
	}
	fmt.Println("packker-client-stub", getVersion())
	return nil
}

func (i *genInput) generate(cmd *cobra.Command, args []string) {

	if len(i.env) == 0 {
		i.env = env
	}

	if i.env == "development" {
		fmt.Println(ui.Warn(fmt.Sprintf("==============================================================\nAsset is packed under %s environment\nPack it under production by enabling 'env' flag in unpackker\n==============================================================\n", i.env)))
	}

	if len(i.assetName) == 0 {
		fmt.Println(ui.Error("Asset cannot be null"))
		os.Exit(1)
	}

	i.version = assetVersion
	i.assetPath = path.Join(i.path, i.assetName, i.version)

	if i.assetExists() {
		fmt.Println(ui.Error(fmt.Sprintf("Asset %s was already created in the location %s\n", i.assetName, i.path)))
		os.Exit(1)
	}

	i.path = i.getPath()
	assests := AssetNames()
	for _, asst := range assests {
		if err := RestoreAssets(i.assetPath, asst); err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
	}
}

func (i *genInput) assetExists() bool {
	if _, direrr := os.Stat(i.assetPath); os.IsNotExist(direrr) {
		return false
	}
	return true
}

func (i *genInput) getPath() string {
	if i.path == "." {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(ui.Error(decode.GetStringOfMessage(err)))
			os.Exit(1)
		}
		return dir
	}
	return path.Dir(i.path)
}

func getVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "v%s", assetVersion)
	if env != "" {
		fmt.Fprintf(&versionString, "-%s", env)
	}

	return versionString.String()
}`

var autoGenMessage = `// ----------------------------------------------------------------------------
//
//     ***     UNPACKKER GENERATED CODE    ***    UNPACKKER GENERATED CODE     ***
//
// ----------------------------------------------------------------------------
//
//     This file was generated automatically by Unpackker.
//     This autogenerated code manual intervene can break the code and Upackker would not understand it.
//
//     Get more information on how Unpackker works.
//     https://github.com/nikhilsbhat/unpackker
//
// ----------------------------------------------------------------------------`

// Generate generates the client stub of asset which could be transported easily and is understood by unpackker itself.
func (i *GenInput) Generate() (string, error) {
	i.getTemplate()

	path, err := i.getPath()
	if err != nil {
		return "", err
	}

	i.Path = path
	i.template = fmt.Sprintf("unpackker-client-stub-%s", i.Package)
	i.AutoGenMessage = autoGenMessage
	if i.clinetStubExists() {
		return "", fmt.Errorf("looks like clinet stub %s was created earlier in the location %s", i.template, i.Path)
	}

	// Generating the ClientStub for asset ex: unpackker-client-stub-demo
	fmt.Println(ui.Info(fmt.Sprintf("ClientStub for asset will be generated under %s\n", i.Path)))
	if err := i.genPackkerClinetStubDir(); err != nil {
		return "", err
	}

	// Generating the required files
	var files = map[string]string{
		"main.go":     i.Path,
		"climeta.go":  fmt.Sprintf("%s/%s", i.Path, i.Package),
		"cli.go":      fmt.Sprintf("%s/%s", i.Path, i.Package),
		"flags.go":    fmt.Sprintf("%s/%s", i.Path, i.Package),
		"register.go": fmt.Sprintf("%s/%s", i.Path, i.Package),
	}

	for key, value := range files {
		if err := i.genPackkerClientStubFiles(key, value); err != nil {
			return "", err
		}
	}
	return i.template, nil
}

func (i *GenInput) getPath() (string, error) {
	if i.Path == "." {
		fmt.Println("Yes")
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return dir, nil
	}
	return i.Path, nil
}

func (i *GenInput) genPackkerClinetStubDir() error {
	pathTgen := filepath.Join(i.Path, i.template)
	fmt.Println(ui.Info(fmt.Sprintf("Clientstub for asset %s does not exists, generating one for you\n", i.template)))
	err := os.MkdirAll(path.Join(pathTgen, i.Package), 0777)
	if err != nil {
		return err
	}
	i.Path = pathTgen
	return nil
}

func (i *GenInput) genPackkerClientStubFiles(name, path string) error {
	file, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return err
	}
	defer file.Close()

	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	var tmpl *template.Template
	switch name {
	case "main.go":
		if len(i.TemplateRaw.RootTemp) != 0 {
			tmpl = template.Must(template.New(name).Parse(i.TemplateRaw.RootTemp))
			if err := tmpl.Execute(file, i); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("template not found for main.go")
	case "cli.go":
		if len(i.TemplateRaw.CliTemp) != 0 {
			tmpl = template.Must(template.New(name).Funcs(funcMap).Parse(i.TemplateRaw.CliTemp))
			if err := tmpl.Execute(file, i); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("template not found for cli.go")
	case "climeta.go":
		if len(i.TemplateRaw.CliMetaTemp) != 0 {
			tmpl = template.Must(template.New(name).Funcs(funcMap).Parse(i.TemplateRaw.CliMetaTemp))
			if err := tmpl.Execute(file, i); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("template not found for climeta.go")
	case "flags.go":
		if len(i.TemplateRaw.FlagsTemp) != 0 {
			tmpl = template.Must(template.New(name).Funcs(funcMap).Parse(i.TemplateRaw.FlagsTemp))
			if err := tmpl.Execute(file, i); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("template not found for flags.go")
	case "register.go":
		if len(i.TemplateRaw.RegisterTemp) != 0 {
			tmpl = template.Must(template.New(name).Funcs(funcMap).Parse(i.TemplateRaw.RegisterTemp))
			if err := tmpl.Execute(file, i); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("template not found for register.go")
	}
	return fmt.Errorf("snap.....!! unable to render the templates, looks like they have issues")
}

// Set the templates to defaults if not specified.
func (i *GenInput) getTemplate() {
	if reflect.DeepEqual(i.TemplateRaw, UnpackkerTemplate{}) {
		i.TemplateRaw.RootTemp = rootTemp
		i.TemplateRaw.CliTemp = cliTemp
		i.TemplateRaw.CliMetaTemp = cliMetaTemp
		i.TemplateRaw.FlagsTemp = flagsTemp
		i.TemplateRaw.RegisterTemp = registerTemp
	}
}

func (i *GenInput) clinetStubExists() bool {
	pathTgen := filepath.Join(i.Path, i.template)
	if _, direrr := os.Stat(pathTgen); os.IsNotExist(direrr) {
		return false
	}
	return true
}
