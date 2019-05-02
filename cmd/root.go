package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ladle",
	Short: "ladle is a runtime for Go based lambdas on localhost",
	Long:  "A local runtime for Go based lambdas with a focus on ease of use.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
