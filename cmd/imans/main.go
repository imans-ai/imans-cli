package main

import (
	"fmt"
	"os"

	"github.com/imans-ai/imans-cli/internal/apperrors"
	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/root"
)

func main() {
	app, err := cli.New(cli.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", apperrors.Format(err))
		os.Exit(apperrors.ExitCode(err))
	}

	cmd := root.New(app)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", apperrors.Format(err))
		os.Exit(apperrors.ExitCode(err))
	}
}
