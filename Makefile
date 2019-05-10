ifeq ($(OS),Windows_NT)
  EXE_EXT := .exe
endif

LADLE_TARGET := build/ladle$(EXE_EXT)
ECHO_TARGET := build/echo$(EXE_EXT)

# Build all the binaries
.PHONY: build
build: $(LADLE_TARGET) $(ECHO_TARGET)

# Test everything
.PHONY: test
test:
	go test -cover -race ./...

# Test verbose in CI
.PHONY: ci-test
ci-test:
	go test -v -race ./...

.PHONY: $(LADLE_TARGET)
$(LADLE_TARGET): 
	go build -o $@

.PHONY: $(ECHO_TARGET)
$(ECHO_TARGET):
	go build -o $@ ./lambdas/echo
