package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// CheckErr prints the msg with the prefix 'Error:' and exits with error code 1. If the msg is nil, it does nothing.
func CheckErr(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\033[1;31m%s\033[0m", fmt.Sprintf("Error: %s\n", err.Error()))
		os.Exit(1)
	}
}

func SilenceCmdErrors(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	for _, subCmd := range cmd.Commands() {
		SilenceCmdErrors(subCmd)
	}
}
