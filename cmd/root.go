package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rpcAddress string

func init() {
	rootCmd.PersistentFlags().StringVarP(&rpcAddress, "rpc-address", "r", "localhost:3000", "RPC invocation Address")
}

var rootCmd = &cobra.Command{
	Use: "ladle",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
