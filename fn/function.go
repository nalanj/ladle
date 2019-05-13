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

	// running notes if the command is being executed
	running bool

	// cmd is the command being executed
	cmd *exec.Cmd

	// out is a reader for the output
	out *io.PipeReader

	// done is a channel to signal on stop
	done chan<- string
}

// Dup duplicates the given function's configuration without duplicating its
// execution state
func Dup(f *Function) *Function {
	return &Function{Name: f.Name, Handler: f.Handler}
}

// Start starts the function's executable. This isn't a function on the function
// struct because the function struct exposes all exported functions over RPC
func Start(f *Function, done chan<- string) error {
	if f.cmd != nil {
		panic("Function instance duplicate start")
	}

	f.done = done

	port, portErr := freePort()
	if portErr != nil {
		return portErr
	}
	f.port = port

	handler := f.Handler
	if runtime.GOOS == "windows" && filepath.Ext(handler) == "" {
		// add .exe on windows
		handler += ".exe"
	}

	f.cmd = exec.Command(handler)
	f.cmd.Env = append(os.Environ(), fmt.Sprintf("_LAMBDA_SERVER_PORT=%d", f.port))

	read, write := io.Pipe()
	f.out = read
	f.cmd.Stdout = write
	f.cmd.Stderr = write

	if runErr := f.cmd.Start(); runErr != nil {
		return runErr
	}
	f.running = true

	go f.readOutput()
	go f.watchHandler(handler)

	// give it up to 10 seconds to actually start
	pinged := false
	start := time.Now()
	for time.Now().Before(start.Add(10 * time.Second)) {
		if f.ping() {
			pinged = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !pinged {
		return fmt.Errorf("Fn %s: could not ping on startup", f.Name)
	}

	log.Printf("Fn %s: Started on port %d\n", f.Name, f.port)
	return nil
}

// Stop stops the function
func Stop(f *Function) error {
	if f.cmd != nil && f.cmd.Process != nil {
		killErr := f.cmd.Process.Kill()
		f.running = false
		f.done <- f.Name
		return killErr
	}

	// it wasn't running anyway
	return nil
}

// watchHandler watches the handler for change and if it changes, stops the
// function
func (f *Function) watchHandler(handler string) {
	mtime := time.Now()

	for f.running == true {
		info, statErr := os.Stat(handler)
		if statErr != nil {
			panic(statErr)
		}

		if info.ModTime().After(mtime) {
			log.Printf("Fn %s: Handler changed, stopping\n", f.Name)
			stopErr := Stop(f)
			if stopErr != nil {
				panic(stopErr)
			}
			break
		}

		time.Sleep(1 * time.Second)
	}
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

// rpcClient returns a new rpc client
func (f *Function) rpcClient() (*rpc.Client, error) {
	return rpc.Dial("tcp", fmt.Sprintf(":%d", f.port))
}

// ping pings the given function and returns true on success, or false if the
// ping failed for any reason
func (f *Function) ping() bool {
	client, clientErr := f.rpcClient()
	if clientErr != nil {
		return false
	}

	pingErr := client.Call("Function.Ping", &messages.PingRequest{}, &messages.PingResponse{})
	if pingErr != nil {
		return false
	}

	return true
}

// Invoke invokes the given function with the given payload
func (f *Function) Invoke(req *messages.InvokeRequest, resp *messages.InvokeResponse) error {
	startTime := time.Now()

	client, clientErr := f.rpcClient()
	if clientErr != nil {
		fmt.Println(clientErr)
		return clientErr
	}

	callErr := client.Call("Function.Invoke", req, resp)
	log.Printf(
		"Fn %s(%s): Invoke (%.3fms)",
		f.Name,
		req.RequestId,
		float64(time.Now().Sub(startTime).Nanoseconds())/1000000,
	)

	return callErr
}

// freePort grabs a free port
func freePort() (int, error) {
	ln, listenErr := net.Listen("tcp", ":0")
	if listenErr != nil {
		return 0, listenErr
	}

	port := ln.Addr().(*net.TCPAddr).Port
	if closeErr := ln.Close(); closeErr != nil {
		return 0, closeErr
	}

	return port, nil
}
