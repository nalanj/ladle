package fn

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

// Function represents a specific function being run
type Function struct {
	// Name is the name of the function
	Name string

	// Handler is the path to the function binary
	Handler string

	// env is the function's environment
	env map[string]string

	// port is the port for the function
	port int

	// cmd is the command being executed
	cmd *exec.Cmd

	// out is a reader for the output
	out *io.PipeReader
}

// Start starts the function's executable
func Start(f *Function) error {
	port, portErr := freePort()
	if portErr != nil {
		return portErr
	}
	f.port = port

	f.cmd = exec.Command(f.Handler)
	f.cmd.Env = append(os.Environ(), fmt.Sprintf("_LAMBDA_SERVER_PORT=%d", f.port))

	read, write := io.Pipe()
	f.out = read
	f.cmd.Stdout = write
	f.cmd.Stderr = write
	go f.readOutput()

	log.Printf("Fn %s: Starting on port %d\n", f.Name, f.port)
	runErr := f.cmd.Start()
	return runErr
}

// readOutput reads output from the output buffer
func (f *Function) readOutput() {
	r := bufio.NewReader(f.out)
	for {
		line, readErr := r.ReadString('\n')
		if readErr != nil {
			fmt.Println(readErr)
			break
		}

		log.Printf("Fn %s: %s", f.Name, line)
	}
}

// Invoke invokes the given function with the given payload
func (f *Function) Invoke(req *messages.InvokeRequest, resp *messages.InvokeResponse) error {
	client, clientErr := rpc.Dial("tcp", fmt.Sprintf("localhost:%d", f.port))
	if clientErr != nil {
		return clientErr
	}

	callErr := client.Call("Function.Invoke", req, resp)
	if callErr != nil {
		return callErr
	}

	return nil
}

// freePort grabs a free port
func freePort() (int, error) {
	ln, listenErr := net.Listen("tcp", "localhost:0")
	if listenErr != nil {
		return 0, listenErr
	}

	port := ln.Addr().(*net.TCPAddr).Port
	if closeErr := ln.Close(); closeErr != nil {
		return 0, closeErr
	}

	return port, nil
}
