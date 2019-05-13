package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/lambda/messages"
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

var runningFunctions map[string]*fn.Function
var runningFunctionsMtx sync.Mutex

func serve() error {
	conf, confErr := config.ParsePath(configPath)
	if confErr != nil {
		return confErr
	}

	runningFunctionsMtx.Lock()
	runningFunctions = make(map[string]*fn.Function)

	fnDone := make(chan string, 20)
	for _, f := range conf.Functions {
		execFn := fn.Dup(f)
		err := fn.Start(execFn, fnDone)
		if err != nil {
			panic(err)
		}

		rpc.RegisterName(execFn.Name, &RPCInvokeWrapper{Name: execFn.Name})
		runningFunctions[execFn.Name] = execFn
	}
	runningFunctionsMtx.Unlock()

	go rpcListener(conf)
	go httpListener(conf, globalInvoker)

	for {
		select {
		case restart := <-fnDone:
			log.Printf("Core: Restarting Fn %s\n", restart)

			runningFunctionsMtx.Lock()
			oldFn, ok := runningFunctions[restart]
			if !ok {
				log.Printf(
					"Core: Attempted restart on inactive function %s\n",
					restart,
				)
				continue
			}

			delete(runningFunctions, restart)

			newFn := fn.Dup(oldFn)
			err := fn.Start(newFn, fnDone)
			if err != nil {
				panic(err)
			}

			runningFunctions[newFn.Name] = newFn
			runningFunctionsMtx.Unlock()
		}
	}
}

func globalInvoker(
	name string,
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error {
	runningFunctionsMtx.Lock()
	runningFunc, ok := runningFunctions[name]
	runningFunctionsMtx.Unlock()

	if !ok {
		return fmt.Errorf("Function %s not running", runningFunc.Name)
	}

	return runningFunc.Invoke(req, resp)
}

// RPCInvokeWrapper exposes Invoke for the given function
type RPCInvokeWrapper struct {
	Name string
}

// Invoke invokes the given function
func (r *RPCInvokeWrapper) Invoke(
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error {
	return globalInvoker(r.Name, req, resp)
}

// rpcListener listens with rpc to the given port and passes messages on
// to the called function
func rpcListener(conf *config.Config) {
	lis, lisErr := net.Listen("tcp", rpcAddress)
	if lisErr != nil {
		panic(lisErr)
	}

	log.Printf("RPC: Listening on %s\n", rpcAddress)
	rpc.Accept(lis)
}

// httpListener starts up a listener that simulates api gateway
func httpListener(conf *config.Config, i fn.Invoker) {
	log.Printf("HTTP: Listening on %s\n", httpAddress)
	http.ListenAndServe(httpAddress, gw.InvokeHandler(conf, i))
}
