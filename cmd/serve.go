package cmd

import (
	"fmt"
	"os"

	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/core"

	"github.com/nalanj/confl"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve functions locally",
	Long: `Serve lambda functions locally. This service must be running for most
			other commands to work`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, confErr := config.ParsePath(configPath)
		if confErr != nil {
			if parseErr, ok := confErr.(*confl.ParseError); ok {
				fmt.Println(parseErr.ErrorWithCode())
			}

			fmt.Println(confErr)
		}

		conf.RPCAddress = rpcAddress
		conf.HTTPAddress = httpAddress

		if err := core.Start(conf); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
