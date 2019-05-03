# Build all the binaries
.PHONY: build
build:
	go build -o build/ladle .
	go build -o build/echo ./lambdas/echo/

# Test everything
.PHONY: test
test:
	go test -race ./...
