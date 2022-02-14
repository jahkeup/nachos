go=go
goimports=go run golang.org/x/tools/cmd/goimports

V=$(if $(CI),-v)

pkg_name=github.com/jahkeup/nachos
cli_pkg=./cmd/nachos
try_pkg=./cmd/nacho-sauce

bin_name=bin-universal
remote_bin_path=/tmp/$(bin_name)

export CGO_ENABLED=0

all: build test

build: bin/nachos_darwin bin/nachos

bin/nachos:
	$(go) build -o $@ $(cli_pkg)

bin/nachos_darwin: bin/nachos_darwin_amd64 bin/nachos_darwin_arm64
	$(go) run $(cli_pkg) -output $@ $^

bin/nachos_darwin_%: $(wildcard *.go ./cmd/nachos/*.go)
	GOARCH=$* GOOS=darwin \
	$(go) build $(V) -o $@ $(cli_pkg)

.PHONY: test
test: test-stubs
	$(go) test $(V) ./...

.PHONY: test-stubs
test-stubs: internal/testing/bins/stub-binary-darwin_amd64
test-stubs: internal/testing/bins/stub-binary-darwin_arm64

internal/testing/bins/stub-binary-%: $(wildcard internal/testing/bins/*.go internal/testing/bins/internal/stub/*.go)
	$(go) generate $(V) ./internal/testing/bins

.PHONY: goimports
goimports:
	$(goimports) -local $(pkg_name) -w .

clean:
	go clean ./...
	rm -rf ./bin
