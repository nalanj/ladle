ifeq ($(OS),Windows_NT)
  EXE_EXT := .exe
endif

LADLE_TARGET := build/ladle$(EXE_EXT)

# Build all the binaries
.PHONY: build
build: $(LADLE_TARGET)
	$(LADLE_TARGET) build
	
# Test everything
.PHONY: test
test:
	go test ./...

# Test verbose in CI
.PHONY: ci-test
ci-test:
	go test -v -race ./...

.PHONY: $(LADLE_TARGET)
$(LADLE_TARGET): 
	go build -o $@
