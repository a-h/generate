PKG := .
CMD := $(PKG)/cmd/schema-generate
BIN := schema-generate

# Build

.PHONY: all clean

all: clean $(BIN)

$(BIN): generator.go jsonschema.go cmd/schema-generate/main.go
	@echo "+ Building $@"
	CGO_ENABLED="0" go build -v -o $@ $(CMD)

clean:
	@echo "+ Cleaning $(PKG)"
	go clean -i $(PKG)/...
	rm -f $(BIN)
	rm -rf test/*_gen

# Test

# generate sources
JSON := $(wildcard test/*.json)
GENERATED_SOURCE := $(patsubst %.json,%_gen/generated.go,$(JSON))
test/%_gen/generated.go: test/%.json 
	@echo "\n+ Generating code for $@"
	@D=$(shell echo $^ | sed 's/.json/_gen/'); \
	[ ! -d $$D ] && mkdir -p $$D || true
	./schema-generate -o $@ -p $(shell echo $^ | sed 's/test\///; s/.json//')  $^

.PHONY: test codecheck fmt lint vet

test: $(BIN) $(GENERATED_SOURCE)
	@echo "\n+ Executing tests for $(PKG)"
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
