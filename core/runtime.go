package core

import (
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/gw"
	"github.com/nalanj/ladle/rpc"
)

var runningFunctions map[string]*FunctionExec
var runningFunctionsMtx sync.Mutex

// StartRuntime starts the runtime
func StartRuntime(conf *config.Config) error {
	runningFunctionsMtx.Lock()
	runningFunctions = make(map[string]*FunctionExec)

	fnDone := make(chan string, 20)
	for _, f := range conf.Functions {
		fnEx, err := StartFunction(conf, f, fnDone)
		if err != nil {
			panic(err)
		}

		rpc.Register(f.Name)
		runningFunctions[f.Name] = fnEx
	}
	runningFunctionsMtx.Unlock()

	go rpc.Listen(conf, globalInvoker)
	go gw.Listener(conf, globalInvoker)

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

			fnEx, err := StartFunction(conf, oldEx.Function, fnDone)
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
