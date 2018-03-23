PKG := github.com/a-h/generate
CMD := $(PKG)/cmd/schema-generate
BIN := schema-generate

# Build

.PHONY: all clean

all: clean $(BIN)

$(BIN):
	@echo "+ Building $@"
	CGO_ENABLED="0" go build -v -o $@ $(CMD)

clean:
	@echo "+ Cleaning $(PKG)"
	go clean -i $(PKG)/...
	rm -f $(BIN)

# Test

.PHONY: test codecheck fmt lint vet

test:
	@echo "+ Executing tests for $(PKG)"
	go test -v -race -cover $(PKG)/...

codecheck: fmt lint vet

fmt:
	@echo "+ go fmt"
	go fmt $(PKG)/...

lint: $(GOPATH)/bin/golint
	@echo "+ go lint"
	golint -min_confidence=0.1 $(PKG)/...

$(GOPATH)/bin/golint:
	go get -v golang.org/x/lint/golint

vet:
	@echo "+ go vet"
	go vet $(PKG)/...
