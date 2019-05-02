package cmd

import (
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   "invoke [function]",
	Short: "Invoke the given function with stdin",
	Long: `
		Invoke runs the given function with the input from stdin.
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, clientErr := rpc.Dial("tcp", "localhost:3000")
		if clientErr != nil {
			fmt.Println(clientErr)
			os.Exit(1)
		}

		payload, readErr := ioutil.ReadAll(os.Stdin)
		if readErr != nil {
			fmt.Println(readErr)
			os.Exit(1)
		}

		req := &messages.InvokeRequest{
			Payload: payload,
		}
		resp := &messages.InvokeResponse{}

		callErr := client.Call(fmt.Sprintf("%s.Invoke", args[0]), req, resp)
		if callErr != nil {
			fmt.Println(callErr)
			os.Exit(1)
		}

		if resp.Error != nil {
			fmt.Printf("Error: %s\n", resp.Error.Message)
			os.Exit(1)
		}

		fmt.Println(string(resp.Payload))
	},
}
