package cmd

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"github.com/nalanj/ladle/fn"
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
		err := serve()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var functions map[string]*fn.Function

func serve() error {
	done := make(chan bool)

	functions = make(map[string]*fn.Function)

	f := &fn.Function{
		Name:    "Hello",
		Handler: "../hello-lambda/hello/hello",
	}

	err := fn.Start(f)
	if err != nil {
		panic(err)
	}

	functions[f.Name] = f

	go invokeListener()
	<-done

	return nil
}

// invokeListener listens with rpc to the given port and passes messages on to the
// called function
func invokeListener() {
	lis, lisErr := net.Listen("tcp", rpcAddress)
	if lisErr != nil {
		panic(lisErr)
	}

	for _, f := range functions {
		rpc.RegisterName(f.Name, f)
	}

	log.Printf("RPC: Listening on %s\n", rpcAddress)
	rpc.Accept(lis)
}
