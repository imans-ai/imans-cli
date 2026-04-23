package flags

import "github.com/spf13/cobra"

const (
	FlagJSON    = "json"
	FlagQuiet   = "quiet"
	FlagNoColor = "no-color"
	FlagProfile = "profile"
	FlagDebug   = "debug"
)

type Options struct {
	JSON    bool
	Quiet   bool
	NoColor bool
	Profile string
	Debug   bool
}

func AddPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(FlagJSON, false, "Output machine-readable JSON")
	cmd.PersistentFlags().Bool(FlagQuiet, false, "Suppress non-essential stderr output")
	cmd.PersistentFlags().Bool(FlagNoColor, false, "Disable color output")
	cmd.PersistentFlags().String(FlagProfile, "", "Use a specific saved profile")
	cmd.PersistentFlags().Bool(FlagDebug, false, "Print safe request diagnostics to stderr")
}

func OptionsFromCommand(cmd *cobra.Command) Options {
	jsonOutput, _ := cmd.Flags().GetBool(FlagJSON)
	quiet, _ := cmd.Flags().GetBool(FlagQuiet)
	noColor, _ := cmd.Flags().GetBool(FlagNoColor)
	profile, _ := cmd.Flags().GetString(FlagProfile)
	debug, _ := cmd.Flags().GetBool(FlagDebug)
	return Options{JSON: jsonOutput, Quiet: quiet, NoColor: noColor, Profile: profile, Debug: debug}
}
