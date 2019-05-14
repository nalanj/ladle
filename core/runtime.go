package core

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"sync"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/fn"
	"github.com/nalanj/ladle/gw"
)

var runningFunctions map[string]*fn.FunctionExec
var runningFunctionsMtx sync.Mutex

// Start starts the runtime
func Start(conf *config.Config) error {
	runningFunctionsMtx.Lock()
	runningFunctions = make(map[string]*fn.FunctionExec)

	fnDone := make(chan string, 20)
	for _, f := range conf.Functions {
		fnEx, err := fn.Start(f, fnDone)
		if err != nil {
			panic(err)
		}

		rpc.RegisterName(f.Name, &RPCInvokeWrapper{Name: f.Name})
		runningFunctions[f.Name] = fnEx
	}
	runningFunctionsMtx.Unlock()

	go rpcListener(conf)
	go httpListener(conf, globalInvoker)

	for {
		select {
		case restart := <-fnDone:
			log.Printf("Core: Restarting Fn %s\n", restart)

			runningFunctionsMtx.Lock()
			oldEx, ok := runningFunctions[restart]
			if !ok {
				log.Printf(
					"Core: Attempted restart on inactive function %s\n",
					restart,
				)
				continue
			}

			delete(runningFunctions, restart)

			fnEx, err := fn.Start(oldEx.Function, fnDone)
			if err != nil {
				panic(err)
			}

			runningFunctions[fnEx.Function.Name] = fnEx
			runningFunctionsMtx.Unlock()
		}
	}
}

// globalInvoker is an invoker based on the runtime function config
func globalInvoker(
	name string,
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error {
	runningFunctionsMtx.Lock()
	fnEx, ok := runningFunctions[name]
	runningFunctionsMtx.Unlock()

	if !ok {
		return fmt.Errorf("Function %s not running", name)
	}

	return fnEx.Invoke(req, resp)
}

// httpListener starts up a listener that simulates api gateway
func httpListener(conf *config.Config, i fn.Invoker) {
	log.Printf("HTTP: Listening on %s\n", conf.HTTPAddress)
	http.ListenAndServe(conf.HTTPAddress, gw.InvokeHandler(conf, i))
}
