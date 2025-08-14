# Makefile for generating protobuf files

.PHONY: config
config:
	protoc --proto_path=. \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:. \
	       conf/conf.proto

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

.PHONY: clean
clean:
	rm -f conf/*.pb.go

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  config  - Generate Go code from proto files"
	@echo "  init    - Install required protobuf tools"
	@echo "  clean   - Remove generated Go files"
	@echo "  help    - Show this help message"