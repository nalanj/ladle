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
	"path/filepath"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

// FunctionExec is a running function
type FunctionExec struct {

	// Function is the function definition being run
	Function *Function

	// port is the port for the function
	port int

	// cmd is the command being executed
	cmd *exec.Cmd

	// done is a channel to signal on stop
	done chan<- string
}

// Start starts the given function and returns a FunctionExec.
func Start(f *Function, done chan<- string) (*FunctionExec, error) {
	fnEx := &FunctionExec{Function: f}
	fnEx.done = done

	port, portErr := freePort()
	if portErr != nil {
		return nil, portErr
	}
	fnEx.port = port

	handler := f.Handler
	if runtime.GOOS == "windows" && filepath.Ext(handler) == "" {
		// add .exe on windows
		handler += ".exe"
	}

	fnEx.cmd = exec.Command(handler)
	fnEx.cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("_LAMBDA_SERVER_PORT=%d", fnEx.port),
	)

	read, write := io.Pipe()
	fnEx.cmd.Stdout = write
	fnEx.cmd.Stderr = write

	if runErr := fnEx.cmd.Start(); runErr != nil {
		return nil, runErr
	}

	go readOutput(f.Name, read)
	go fnEx.watchHandler(handler)

	// give it up to 10 seconds to actually start
	pinged := false
	start := time.Now()
	for time.Now().Before(start.Add(10 * time.Second)) {
		if fnEx.ping() {
			pinged = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !pinged {
		return nil, fmt.Errorf("Fn %s: could not ping on startup", f.Name)
	}

	log.Printf("Fn %s: Started on port %d\n", f.Name, fnEx.port)
	return fnEx, nil
}

// Stop stops the function
func Stop(fnEx *FunctionExec) error {
	if fnEx.cmd != nil && fnEx.cmd.Process != nil {
		killErr := fnEx.cmd.Process.Kill()
		fnEx.done <- fnEx.Function.Name
		return killErr
	}

	// it wasn't running anyway
	return nil
}

// watchHandler watches the handler for change and if it changes, stops the
// function
func (fnEx *FunctionExec) watchHandler(handler string) {
	mtime := time.Now()

	for {
		info, statErr := os.Stat(handler)
		if statErr != nil {
			panic(statErr)
		}

		if info.ModTime().After(mtime) {
			log.Printf(
				"Fn %s: Handler changed, stopping\n",
				fnEx.Function.Name,
			)
			stopErr := Stop(fnEx)
			if stopErr != nil {
				panic(stopErr)
			}
			break
		}

		time.Sleep(1 * time.Second)
	}
}

// readOutput reads output from the output buffer
func readOutput(name string, out io.ReadCloser) {
	r := bufio.NewReader(out)
	for {
		line, readErr := r.ReadString('\n')
		if readErr != nil {
			fmt.Println(readErr)
			break
		}

		log.Printf("Fn %s: %s", name, line)
	}

	out.Close()
}

// rpcClient returns a new rpc client
func (fnEx *FunctionExec) rpcClient() (*rpc.Client, error) {
	return rpc.Dial("tcp", fmt.Sprintf(":%d", fnEx.port))
}

// ping pings the given function and returns true on success, or false if the
// ping failed for any reason
func (fnEx *FunctionExec) ping() bool {
	client, clientErr := fnEx.rpcClient()
	if clientErr != nil {
		return false
	}

	pingErr := client.Call(
		"Function.Ping",
		&messages.PingRequest{},
		&messages.PingResponse{},
	)

	if pingErr != nil {
		return false
	}

	return true
}

// Invoke invokes the given function with the given payload
func (fnEx *FunctionExec) Invoke(
	req *messages.InvokeRequest,
	resp *messages.InvokeResponse,
) error {
	startTime := time.Now()

	client, clientErr := fnEx.rpcClient()
	if clientErr != nil {
		fmt.Println(clientErr)
		return clientErr
	}

	callErr := client.Call("Function.Invoke", req, resp)
	log.Printf(
		"Fn %s(%s): Invoke (%.3fms)",
		fnEx.Function.Name,
		req.RequestId,
		float64(time.Now().Sub(startTime).Nanoseconds())/1000000,
	)

	return callErr
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
