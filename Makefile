GOROOT := $(shell go env GOROOT)
DIRS := \
	$(TOOL_BIN_DIR) \
	$(TOOL_TEMP_DIR)
export PATH := $(TOOL_BIN_DIR):$(PATH)
include $(wildcard $(TOOL_DIR)/*.mk)

PROTOS := \
	**/*.proto

$(DIRS):
	mkdir -p $@

PHONY+= init
init: 
	git submodule init
	git submodule update

PHONY+= tools
tools: $(DIRS) $(PROTOC)
	@go install \
		google.golang.org/protobuf/cmd/protoc-gen-go 

PHONY += protobuf
protobuf:
	@for d in $$(find "crypto" -type f -name "*.proto"); do		\
		protoc -I$(GOPATH)/src --go_out=$(GOPATH)/src $(CURDIR)/$$d; \
	done;

PHONY += tss
tss:
	go build -o cmd/tss main.go

.PHONY: $(PHONY)
