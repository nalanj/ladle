package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rpcAddress string
var httpAddress string

func init() {
	rootCmd.PersistentFlags().StringVarP(&rpcAddress, "rpc-address", "r", "localhost:3000", "RPC invocation Address")
	rootCmd.PersistentFlags().StringVarP(&httpAddress, "http-address", "a", "localhost:3001", "API Gateway Address")
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
