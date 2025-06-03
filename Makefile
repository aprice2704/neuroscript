# Makefile for NeuroScript Project

# Go variables
GO := go
GOFLAGS :=

# Directories
ROOT_DIR := $(shell pwd)
BIN_DIR := $(ROOT_DIR)/bin # Kept for 'clean' and potential ad-hoc local builds
CMDS_DIR := $(ROOT_DIR)/cmd
PKG_DIR := $(ROOT_DIR)/pkg
VSCODE_EXT_DIR := $(ROOT_DIR)/vscode-neuroscript
VSCODE_SERVER_DIR := $(VSCODE_EXT_DIR)/server
INDEX_OUTPUT_DIR := $(PKG_DIR)/codebase-indices # For goindexer output

# Project Go module files (dependencies for all Go builds)
PROJECT_GO_MOD := $(ROOT_DIR)/go.mod
PROJECT_GO_SUM := $(ROOT_DIR)/go.sum

# ANTLR variables
ANTLR_JAR := $(PKG_DIR)/antlr4-4.13.2-complete.jar
G4_FILE := $(PKG_DIR)/core/NeuroScript.g4
G4_TXT_FILE := $(PKG_DIR)/core/NeuroScript.g4.txt
ANTLR_OUTPUT_DIR := $(PKG_DIR)/core/generated
ANTLR_STAMP_FILE := $(ANTLR_OUTPUT_DIR)/.antlr-generated-stamp

# VSCode Extension Stamp File
VSCODE_BUILD_STAMP := $(VSCODE_EXT_DIR)/.vsix-built-stamp

# Find all .go files in the pkg/ directory to use as dependencies for commands
ALL_PKG_GO_FILES := $(shell find $(PKG_DIR) -name '*.go')

# Versioning
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
MODULE_PATH := $(shell go list -m)
# Initial ANTLR_GRAMMAR_VERSION for immediate use (e.g., echo). Will be re-evaluated in rules using LDFLAGS.
ANTLR_GRAMMAR_VERSION_INIT := $(shell if [ -f "$(G4_FILE)" ]; then grep '// Version:' $(G4_FILE) | awk '{print $$NF}'; else echo "g4_not_found"; fi)

# LDFLAGS template base - specific version part will be completed within rules
LDFLAGS_BASE := -X main.AppVersion=$(GIT_VERSION) -X $(MODULE_PATH)/pkg/core.GrammarVersion=

# Tools
VSCE := vsce # Assumes vsce is installed and in PATH
GOINDEXER := goindexer # Assumes goindexer is installed and in PATH

# --- Command Definitions ---
# Find all subdirectories in cmd/ that are likely Go commands.
# Exclude 'nslsp' because it has special handling for the VSCode extension.
# Exclude 'nssync' because its main file is 'main.go.txt'.
COMMAND_SOURCE_DIRS := $(filter-out $(CMDS_DIR)/nslsp $(CMDS_DIR)/nssync, $(shell find $(CMDS_DIR) -mindepth 1 -maxdepth 1 -type d))
CMD_NAMES := $(notdir $(COMMAND_SOURCE_DIRS))


# --- PHONY Targets (targets that don't represent actual files) ---
.PHONY: all build install test clean help generate-antlr always-build-index

# --- Default Target ---
all: build

# --- Target to always build the codebase index ---
always-build-index:
	@echo "Building codebase index into $(INDEX_OUTPUT_DIR)..."
	@mkdir -p $(INDEX_OUTPUT_DIR)
	$(GOINDEXER) -root . -output $(INDEX_OUTPUT_DIR)

# --- Main Build Target ---
# This target installs all standard Go commands using 'go install',
# and prepares/builds the VSCode extension, ANTLR files, .g4.txt, and codebase index.
build: always-build-index $(ANTLR_STAMP_FILE) $(G4_TXT_FILE) $(VSCODE_SERVER_DIR)/nslsp_executable $(VSCODE_BUILD_STAMP)
	$(eval CURRENT_ANTLR_GRAMMAR_VERSION := $(shell if [ -f "$(G4_FILE)" ]; then grep '// Version:' $(G4_FILE) | awk '{print $$NF}'; else echo "g4_not_found"; fi))
	$(eval CURRENT_LDFLAGS := -ldflags="$(LDFLAGS_BASE)$(CURRENT_ANTLR_GRAMMAR_VERSION)")
	@echo "Installing Go commands from ./cmd/... using 'go install' with LDFLAGS..."
	$(GO) install $(GOFLAGS) $(CURRENT_LDFLAGS) ./cmd/...
	@echo ""
	@echo "NeuroScript Build Complete!"
	@echo "--------------------------------------------------"
	@echo "App Version: $(GIT_VERSION), Grammar Version: $(CURRENT_ANTLR_GRAMMAR_VERSION)"
	@echo "Commands installed to GOBIN or GOPATH/bin (via go install)"
	@echo "Codebase index updated in: $(INDEX_OUTPUT_DIR)"
	@echo "nslsp_executable for VSCode at: $(VSCODE_SERVER_DIR)/nslsp_executable"
	@echo "VSCode extension package: $(VSCODE_EXT_DIR)/*.vsix (if built)"
	@echo "VSCode build stamp: $(VSCODE_BUILD_STAMP)"
	@echo "NeuroScript.g4.txt updated at: $(G4_TXT_FILE)"
	@echo "ANTLR parser files generated in: $(ANTLR_OUTPUT_DIR)/"
	@echo "--------------------------------------------------"

# --- ANTLR Generation & G4 Text File ---
generate-antlr: $(ANTLR_STAMP_FILE) $(G4_TXT_FILE)

$(G4_TXT_FILE): $(G4_FILE)
	@echo "Copying $(G4_FILE) to $(G4_TXT_FILE)..."
	cp $< $@

$(ANTLR_STAMP_FILE): $(G4_FILE) $(ANTLR_JAR)
	@echo "Generating ANTLR parser files from $(G4_FILE)..."
	@mkdir -p $(ANTLR_OUTPUT_DIR)
	java -jar $(ANTLR_JAR) -Dlanguage=Go -o $(ANTLR_OUTPUT_DIR) -visitor -listener -package core $(G4_FILE)
	@touch $@
	@echo "ANTLR generation complete. Stamp file: $@"


# --- Rules for nslsp (NeuroScript Language Server Protocol) ---
NSLSP_EXEC_SRC_PATH := $(CMDS_DIR)/nslsp/nslsp_executable

$(NSLSP_EXEC_SRC_PATH): $(shell find $(CMDS_DIR)/nslsp -name '*.go') $(ALL_PKG_GO_FILES) $(PROJECT_GO_MOD) $(PROJECT_GO_SUM) $(ANTLR_STAMP_FILE)
	$(eval CURRENT_ANTLR_GRAMMAR_VERSION := $(shell if [ -f "$(G4_FILE)" ]; then grep '// Version:' $(G4_FILE) | awk '{print $$NF}'; else echo "g4_not_found"; fi))
	$(eval CURRENT_LDFLAGS := -ldflags="$(LDFLAGS_BASE)$(CURRENT_ANTLR_GRAMMAR_VERSION)")
	@echo "Building nslsp executable: $(NSLSP_EXEC_SRC_PATH)..."
	cd $(CMDS_DIR)/nslsp && $(GO) build $(GOFLAGS) $(CURRENT_LDFLAGS) -o nslsp_executable main.go

$(VSCODE_SERVER_DIR)/nslsp_executable: $(NSLSP_EXEC_SRC_PATH)
	@echo "Copying nslsp_executable to $(VSCODE_SERVER_DIR)/ and setting permissions..."
	@mkdir -p $(VSCODE_SERVER_DIR)
	cp $(NSLSP_EXEC_SRC_PATH) $@
	chmod +x $@


# --- Rule for Building VSCode Extension ---
$(VSCODE_BUILD_STAMP): $(VSCODE_SERVER_DIR)/nslsp_executable \
						$(shell find $(VSCODE_EXT_DIR) -maxdepth 1 -name 'package.json') \
						$(shell find $(VSCODE_EXT_DIR) \( -name '*.json' -o -name '*.js' -o -name 'package.nls.json' -o -name '*.ts' -o -name '*.tsx' -o -name '*.md' \) -type f ! -path '$(VSCODE_EXT_DIR)/server/*' ! -path '$(VSCODE_EXT_DIR)/node_modules/*')
	@echo "Packaging VSCode extension in $(VSCODE_EXT_DIR)..."
	cd $(VSCODE_EXT_DIR) && $(VSCE) package
	@echo "VSCode extension packaged successfully. Updating stamp."
	@touch $@


# --- Installation Target ---
# Installs commands and ensures ANTLR and codebase index are up-to-date.
install: always-build-index $(ANTLR_STAMP_FILE) $(PROJECT_GO_MOD) $(PROJECT_GO_SUM)
	$(eval CURRENT_ANTLR_GRAMMAR_VERSION := $(shell if [ -f "$(G4_FILE)" ]; then grep '// Version:' $(G4_FILE) | awk '{print $$NF}'; else echo "g4_not_found"; fi))
	$(eval CURRENT_LDFLAGS := -ldflags="$(LDFLAGS_BASE)$(CURRENT_ANTLR_GRAMMAR_VERSION)")
	@echo "Installing all commands from ./cmd/... using 'go install' with version flags"
	$(GO) install $(GOFLAGS) $(CURRENT_LDFLAGS) ./cmd/...
	@echo "Installation complete. Binaries are in your GOBIN or GOPATH/bin."


# --- Test Target ---
test:
	@echo "Running Go tests for all packages (./...)..."
	$(GO) test ./...


# --- Clean Target ---
clean:
	@echo "Cleaning build artifacts..."
	-rm -rf $(BIN_DIR) # Clean local bin dir in case of old or manual builds
	-rm -f $(CMDS_DIR)/nslsp/nslsp_executable
	-rm -f $(VSCODE_EXT_DIR)/*.vsix
	-rm -f $(VSCODE_SERVER_DIR)/nslsp_executable
	-rm -f $(G4_TXT_FILE)
	-rm -rf $(ANTLR_OUTPUT_DIR)
	-rm -f $(VSCODE_BUILD_STAMP)
	-rm -rf $(INDEX_OUTPUT_DIR) # Clean the codebase index
	@echo "Clean complete."
	@echo "Consider running 'go clean -cache' or 'go clean -modcache' for a deeper Go clean."


# --- Help Target ---
help:
	@echo "NeuroScript Project Makefile"
	@echo "----------------------------"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Core Targets:"
	@echo "  all                      - Build commands, VSCode extension, ANTLR, index (default)"
	@echo "  build                    - Alias for 'all'"
	@echo "  install                  - Install Go commands (to GOBIN/GOPATH), build ANTLR & index"
	@echo "  test                     - Run Go tests (go test ./...)"
	@echo "  clean                    - Remove build artifacts, generated files, and index"
	@echo ""
	@echo "Component Build Targets:"
	@echo "  --- NSLSP & VSCode Extension ---"
	@echo "  $(NSLSP_EXEC_SRC_PATH)   - Build nslsp executable in ./cmd/nslsp/"
	@echo "  $(VSCODE_SERVER_DIR)/nslsp_executable - Copy nslsp to VSCode server dir"
	@echo "  $(VSCODE_BUILD_STAMP)      - Package the VSCode extension (creates .vsix and stamp)"
	@echo ""
	@echo "  --- ANTLR / Code Generation ---"
	@echo "  generate-antlr           - Copy .g4 to .g4.txt and generate ANTLR parser files"
	@echo "  $(G4_TXT_FILE)           - Copy NeuroScript.g4 to NeuroScript.g4.txt"
	@echo "  $(ANTLR_STAMP_FILE)      - Generate Go parser files from NeuroScript.g4"
	@echo ""
	@echo "  --- Indexing ---"
	@echo "  always-build-index       - Force regeneration of the codebase index"
	@echo ""
	@echo "Other Targets:"
	@echo "  help                     - Show this help message"