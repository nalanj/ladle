package rpc

import (
	"log"
	"net"
	"net/rpc"

	"github.com/nalanj/ladle/config"
)

// Register registers a function name for RPC
func Register(name string) {
	rpc.RegisterName(name, &InvokeWrapper{Name: name})
}

// Listen listens with rpc to the given port and passes messages on
// to the called function
func Listen(conf *config.Config, i Invoker) {
	globalInvoker = i

	lis, lisErr := net.Listen("tcp", conf.RPCAddress)
	if lisErr != nil {
		panic(lisErr)
	}

	log.Printf("RPC: Listening on %s\n", conf.RPCAddress)
	rpc.Accept(lis)
}
