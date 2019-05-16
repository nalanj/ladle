package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nalanj/confl"
	"github.com/nalanj/ladle/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds all lambdas for local execution",
	Long: `
		Build wraps go build as a shortcut for local builds. The executable
		is output to .ladle.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, confErr := config.ParsePath(configPath)
		if confErr != nil {
			if parseErr, ok := confErr.(*confl.ParseError); ok {
				fmt.Println(parseErr.ErrorWithCode())
			}

			fmt.Println(confErr)
			os.Exit(-1)
		}

		if err := conf.EnsureRuntimeDir(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		for funcName, function := range conf.Functions {
			if err := build(conf, funcName, function); err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		}
	},
}

// build runs go build against the given function definition
func build(conf *config.Config, name string, function *config.Function) error {
	args := []string{
		"build",
		"-o",
		conf.FunctionExecutable(function),
		function.Package,
	}
	fmt.Printf("Fn %s: %s %s\n", name, "go", strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
