BUILD_DIR := $(CURDIR)/build
BIN_DIR := $(BUILD_DIR)/bin
BINARY_NAME := "grpc-gateway-experiments"

PROTOC = $(BIN_DIR)/protoc
PROTOC_VERSION = 21.12

PROTOC_GEN_GO = $(BIN_DIR)/protoc-gen-go
PROTOC_GEN_GO_VERSION = latest

PROTOC_GEN_GO_GRPC = $(BIN_DIR)/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION = latest

PROTOC_GEN_GRPC_GATEWAY = $(BIN_DIR)/protoc-gen-grpc-gateway
PROTOC_GEN_GRPC_GATEWAY_VERSION = latest

PROTOC_GEN_DOC = $(BIN_DIR)/protoc-gen-doc
PROTOC_GEN_DOC_VERSION = latest

PROTOC_GEN_OPENAPIV2 = $(BIN_DIR)/protoc-gen-openapiv2
PROTOC_GEN_OPENAPIV2_VERSION = latest

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLANGCI_LINT_VERSION := v1.51.1

OS := $(shell uname -s)

.DEFAULT_GOAL := help

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

init:
	@mkdir -p "$(BUILD_DIR)" "$(BIN_DIR)"

##@ Build

.PHONY: build
build: init ## Build and install the binary.
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

.PHONY: test
test: init build ## Execute unit tests.
	@echo "➡️Running tests..."
	@go test ./...

.PHONY: run
run: init build ## Locally run the server.
	@echo "➡️Launching server..."
	@go run ./main.go

.PHONY: lint
lint: init $(GOLANGCI_LINT) ## Lint the go code.
	@$(GOLANGCI_LINT) run .

.PHONY: generate
generate: init $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_DOC) $(PROTOC_GEN_OPENAPIV2) ## Parse proto/* files and generate output files
	@mkdir -p generated/
	@rm -Rf generated/*.pb*go
	@echo "➡️Generating Go code..."
	@$(PROTOC) \
		--plugin protoc-gen-go=$(PROTOC_GEN_GO) \
		--go_out="./generated" \
		--go_opt=paths=source_relative \
		-I=./proto:$(BUILD_DIR)/include \
		$(shell find ./proto | grep -v google/ | grep -E '\.proto')
	@echo "➡️Generating gRPC code..."
	@$(PROTOC) \
		--plugin protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
		--go-grpc_out="./generated" \
		--go-grpc_opt=paths=source_relative \
		-I=./proto:$(BUILD_DIR)/include \
		$(shell find ./proto | grep -v google/ | grep -E '\.proto')
	@echo "➡️Generating gateway code..."
	@./scripts/build_gateways.sh "$(PROTOC)" "./proto:$(BUILD_DIR)/include" "$(PROTOC_GEN_GRPC_GATEWAY)" "./generated" "./proto"
	cd ./generated && go mod tidy

.PHONY: clean
clean: ## Delete the build directory and generated files.
	@rm -rf $(BUILD_DIR) generated/*pb*

$(PROTOC):
	@./scripts/curl_protoc_$(OS).sh $(PROTOC_VERSION)
	@unzip -q -o /tmp/protoc.zip -d /tmp/protoc
	@cp -f /tmp/protoc/bin/protoc $(PROTOC)
	@cp -rf /tmp/protoc/include $(BUILD_DIR)/include
	@rm -rf /tmp/protoc*

$(PROTOC_GEN_GO):
	@$(call go-install-tool,$(PROTOC_GEN_GO),google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION))

$(PROTOC_GEN_GO_GRPC):
	@$(call go-install-tool,$(PROTOC_GEN_GO_GRPC),google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION))

$(PROTOC_GEN_GRPC_GATEWAY):
	@$(call go-install-tool,$(PROTOC_GEN_GRPC_GATEWAY),github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION))

$(PROTOC_GEN_DOC):
	@$(call go-install-tool,$(PROTOC_GEN_DOC),github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@$(PROTOC_GEN_DOC_VERSION))

$(PROTOC_GEN_OPENAPIV2):
	@$(call go-install-tool,$(PROTOC_GEN_OPENAPIV2),github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(PROTOC_GEN_OPENAPIV2_VERSION))

$(GOLANGCI_LINT):
	@$(call install-golangci-lint-version, $(GOLANGCI_LINT), $(GOLANGCI_LINT_VERSION), $(BIN_DIR))

# go-install-tool will 'go install' any package $2 and install it to $1.
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp 2>/dev/null ;\
echo "Downloading $(2)" ;\
GOBIN=$(BIN_DIR) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# install-golangci-lint-version will check the installed golangci-lint (located in $1) version. If version does not equal $2, then install in $3
define install-golangci-lint-version
@LINT_RESULT=$$($(1) --version 2>/dev/null);\
LINT_VERSION=$$(echo "$$LINT_RESULT" | cut -d ' ' -f 4); \
if [ "$$LINT_RESULT" != "0" ] || [ "$$LINT_VERSION" != "v$(2)" ]; then \
	echo "➡️Installing locally golangci-lint $(2)"; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(3) $(2); \
fi
endef
