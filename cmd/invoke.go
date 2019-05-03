package cmd

import (
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   "invoke [function] [payload-path]",
	Short: "Invoke the given function with stdin",
	Long: `
		Invoke runs the given function with the input from stdin.
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client, clientErr := rpc.Dial("tcp", rpcAddress)
		if clientErr != nil {
			fmt.Println(clientErr)
			os.Exit(1)
		}

		info, statErr := os.Stdin.Stat()
		if statErr != nil {
			fmt.Println(statErr)
			os.Exit(1)
		}

		var payload []byte
		if info.Mode()&os.ModeNamedPipe != 0 {
			var readErr error
			payload, readErr = ioutil.ReadAll(os.Stdin)
			if readErr != nil {
				fmt.Println(readErr)
				os.Exit(1)
			}
		} else {
			if args[1] == "" {
				fmt.Println("No payload specified")
				os.Exit(1)
			}

			payloadFile, openErr := os.Open(args[1])
			if openErr != nil {
				fmt.Println(openErr)
				os.Exit(1)
			}

			var readErr error
			payload, readErr = ioutil.ReadAll(payloadFile)
			if readErr != nil {
				fmt.Println(readErr)
				os.Exit(1)
			}
		}

		req := &messages.InvokeRequest{
			RequestId: uuid.Must(uuid.NewV4()).String(),
			Payload:   payload,
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
