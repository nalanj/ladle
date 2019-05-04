# Build all the binaries
.PHONY: build
build: build/ladle build/echo

# Test everything
.PHONY: test
test:
	go test -race ./...

.PHONY: build/ladle
build/ladle:
	go build -o $@ .

.PHONY: build/echo
build/echo:
	go build -o $@ ./lambdas/echo
