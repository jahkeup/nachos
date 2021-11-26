go=go
goimports=goimports

V=$(if $(CI),-v)

pkg_name=github.com/jahkeup/nachos
run_pkg=./cmd/nachos
try_pkg=./cmd/nacho-sauce

bin_name=bin-universal
remote_bin_path=/tmp/$(bin_name)

export CGO_ENABLED=0

test:
	$(go) test $(V) ./...

goimports:
	$(goimports) -local $(pkg_name) -w .

show: bin-x86_64 bin-arm64 bin-universal
	file $^

try: $(bin_name)
	$(if $(remote_host),,$(error define remote_host to run))
	cat $(bin_name) |\
		ssh "$(remote_host)" -- \
			bash -c ':; rm -f $(remote_bin_path) ;\
					 cat > $(remote_bin_path) ;\
					 chmod 755 $(remote_bin_path) ;\
					 set -x; file $(remote_bin_path) ;\
					 $(remote_bin_path) ;\
					 echo $$?'

$(bin_name): bin-x86_64 bin-arm64
	$(go) run $(V) $(run_pkg) $@

bin-x86_64: $(wildcard *.go) $(wildcard cmd/*/*.go) Makefile
	GOOS=darwin GOARCH=amd64 $(go) build $(V) -o $@ $(try_pkg)

bin-arm64: $(wildcard *.go) $(wildcard cmd/*/*.go) Makefile
	GOOS=darwin GOARCH=arm64 $(go) build $(V) -o $@ $(try_pkg)
