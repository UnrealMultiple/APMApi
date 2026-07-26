cur_dir=$(CURDIR)
idl_dir=$(cur_dir)/idl

.PHONY: hz-gen-api
hz-gen-api:
	@hz update -I idl -idl $(idl_dir)/plugin/plugin.thrift
	@hz update -I idl -idl $(idl_dir)/legacy/legacy.thrift
	@go mod tidy
	@echo '[INFO] Code Generation is Done!'

.PHONY: build
build:
	@go build -o output/apm-api .

.PHONY: build-linux
build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o output/apm-api .

.PHONY: test
test:
	@go test ./...

.PHONY: vet
vet:
	@go vet ./...
