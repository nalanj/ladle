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

var runningFunctions map[string]*fn.Function
var runningFunctionsMtx sync.Mutex

// Start starts the runtime
func Start(conf *config.Config) error {
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

// httpListener starts up a listener that simulates api gateway
func httpListener(conf *config.Config, i fn.Invoker) {
	log.Printf("HTTP: Listening on %s\n", conf.HTTPAddress)
	http.ListenAndServe(conf.HTTPAddress, gw.InvokeHandler(conf, i))
}
