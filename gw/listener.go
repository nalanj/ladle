package gw

import (
	"log"
	"net/http"

	"github.com/nalanj/ladle/config"
	"github.com/nalanj/ladle/rpc"
)

// Listener starts up a listener that simulates api gateway
func Listener(conf *config.Config, i rpc.Invoker) {
	log.Printf("HTTP: Listening on %s\n", conf.HTTPAddress)
	http.ListenAndServe(conf.HTTPAddress, InvokeHandler(conf, i))
}
