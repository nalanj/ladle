package core

import (
	"log"
	"net"
	"net/rpc"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/nalanj/ladle/config"
)

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
	lis, lisErr := net.Listen("tcp", conf.RPCAddress)
	if lisErr != nil {
		panic(lisErr)
	}

	log.Printf("RPC: Listening on %s\n", conf.RPCAddress)
	rpc.Accept(lis)
}
