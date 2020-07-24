package cli

import (
	"fmt"

	"github.com/nikhilsbhat/unpackker/packer"
	"github.com/nikhilsbhat/unpackker/version"
	"github.com/spf13/cobra"
)

var (
	cmds    map[string]*cobra.Command
	unpcker packer.UnpackkerInput
	// genin   gen.GenInput
)

type confcmds struct {
	commands []*cobra.Command
}

// SetUnpackkerCmds helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetUnpackkerCmds() *cobra.Command {
	cmd := getUnpackkerCmds()
	return cmd
}

func getUnpackkerCmds() *cobra.Command {

	var unpackkerCmd = &cobra.Command{
		Use:   "unpackker [command]",
		Short: "Command creates and ship the asset",
		Long:  `unpackker helps user to pack the asset that could be shipped later.`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  cm.echoUnpackker,
	}
	unpackkerCmd.SetUsageTemplate(getUsageTemplate())

	var setCmd = &cobra.Command{
		Use:          "generate [flags]",
		Short:        "Command to generate package of the specified asset",
		Long:         `This will help user to generate package of the specified asset.`,
		Run:          unpcker.Packer,
		SilenceUsage: true,
	}

	// fetching "version" will be done here.
	var versionCmd = &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of unpackker installed",
		Long:  `This will help user to find what version of unpackker he/she installed in her machine.`,
		RunE:  versionConfig,
	}

	unpackkerCmd.AddCommand(setCmd)
	unpackkerCmd.AddCommand(versionCmd)
	registerFlags(unpackkerCmd)
	return unpackkerCmd
}

func (cm *cliMeta) echoUnpackker(cmd *cobra.Command, args []string) error {
	if err := cmd.Usage(); err != nil {
		return err
	}
	return nil
}

func versionConfig(cmd *cobra.Command, args []string) error {
	fmt.Println("unpackker", version.GetVersion())
	return nil
}

// This function will return the custom template for usage function,
// only functions/methods inside this package can call this.

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}"
{{printf "\n"}}`
}
