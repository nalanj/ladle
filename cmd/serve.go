package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda/messages"
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
	go httpListener()
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

// httpListener starts up a listener that simulates api gateway
func httpListener() {
	log.Printf("HTTP: Listening on %s\n", httpAddress)
	http.ListenAndServe(httpAddress, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		f := functions["Hello"]

		body, bodyErr := ioutil.ReadAll(r.Body)
		if bodyErr != nil {
			log.Printf("HTTP: Error: %s\n", bodyErr)
			w.WriteHeader(http.StatusInternalServerError)
		}

		gwR := events.APIGatewayProxyRequest{
			Resource:        "",
			Path:            r.URL.Path,
			HTTPMethod:      r.Method,
			Headers:         make(map[string]string),
			Body:            string(body),
			IsBase64Encoded: false,
			RequestContext:  events.APIGatewayProxyRequestContext{},
		}

		payload, marshalErr := json.Marshal(gwR)
		if marshalErr != nil {
			log.Printf("HTTP: Error: %s\n", marshalErr)
			w.WriteHeader(http.StatusInternalServerError)
		}

		req := &messages.InvokeRequest{Payload: payload}
		resp := &messages.InvokeResponse{}
		invokeErr := f.Invoke(req, resp)
		if invokeErr != nil {
			log.Printf("HTTP: Error: %s\n", invokeErr)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if resp.Error != nil {
			log.Printf("HTTP: Error: %s\n", resp.Error.Message)
			w.WriteHeader(http.StatusInternalServerError)
		}

		var gwResp events.APIGatewayProxyResponse
		if unmarshalErr := json.Unmarshal(resp.Payload, &gwResp); unmarshalErr != nil {
			log.Printf("HTTP: Error: %s\n", unmarshalErr)
			w.WriteHeader(http.StatusInternalServerError)
		}

		for key, val := range gwResp.Headers {
			w.Header().Add(key, val)
		}
		w.WriteHeader(gwResp.StatusCode)
		w.Write([]byte(gwResp.Body))
	}))
}
