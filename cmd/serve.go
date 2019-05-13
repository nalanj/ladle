package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/nalanj/confl"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
	"github.com/nalanj/ladle/gw"
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
			if parseErr, ok := err.(*confl.ParseError); ok {
				fmt.Println(parseErr.ErrorWithCode())
			} else {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	},
}

var functions map[string]*fn.Function

func serve() error {
	done := make(chan bool)

	conf, confErr := config.ParsePath(configPath)
	if confErr != nil {
		return confErr
	}

	fnDone := make(chan string)
	for _, f := range conf.Functions {
		err := fn.Start(f, fnDone)
		if err != nil {
			panic(err)
		}
	}

	go invokeListener(conf)
	go httpListener(conf)
	<-done

	return nil
}

// invokeListener listens with rpc to the given port and passes messages on to the
// called function
func invokeListener(conf *config.Config) {
	lis, lisErr := net.Listen("tcp", rpcAddress)
	if lisErr != nil {
		panic(lisErr)
	}

	for _, f := range conf.Functions {
		rpc.RegisterName(f.Name, f)
	}

	log.Printf("RPC: Listening on %s\n", rpcAddress)
	rpc.Accept(lis)
}

// httpListener starts up a listener that simulates api gateway
func httpListener(conf *config.Config) {
	log.Printf("HTTP: Listening on %s\n", httpAddress)
	http.ListenAndServe(httpAddress, gw.InvokeHandler(conf))
}
