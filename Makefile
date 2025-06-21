# Makefile for NeuroScript Project

# Go variables
GO := go
GOFLAGS :=
# Define the canonical installation directory for Go binaries
BIN_INSTALL_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

# Directories
ROOT_DIR := $(shell pwd)
CMDS_DIR := $(ROOT_DIR)/cmd
PKG_DIR := $(ROOT_DIR)/pkg
VSCODE_EXT_DIR := $(ROOT_DIR)/vscode-neuroscript
VSCODE_SERVER_DIR := $(VSCODE_EXT_DIR)/server
VIM_PLUGIN_DIR := $(ROOT_DIR)/vim-neuroscript
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

# Doc file for GoLand setup
GOLAND_SETUP_DOC := $(ROOT_DIR)/GOLAND_SETUP.md

# Find all .go files in the pkg/ directory to use as dependencies for commands
ALL_PKG_GO_FILES := $(shell find $(PKG_DIR) -name '*.go')

# Versioning
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
MODULE_PATH := $(shell go list -m)
LDFLAGS_BASE := -X main.AppVersion=$(GIT_VERSION) -X $(MODULE_PATH)/pkg/core.GrammarVersion=

# Tools
VSCE := vsce
GOINDEXER := goindexer
SYNTAX_CHECKER := $(BIN_INSTALL_DIR)/syntax-check

# --- Command Definitions ---
COMMAND_SOURCE_DIRS := $(shell find $(CMDS_DIR) -mindepth 1 -maxdepth 1 -type d ! -name 'nssync')
CMD_NAMES := $(notdir $(COMMAND_SOURCE_DIRS))
CMD_BINS := $(addprefix $(BIN_INSTALL_DIR)/, $(CMD_NAMES))

# --- PHONY Targets ---
.PHONY: all build install test clean help generate-antlr always-build-index syntax-check setup-nvim setup-goland

# --- Default Target ---
all: build

# --- Main Build Target ---
build: syntax-check install always-build-index $(G4_TXT_FILE) $(VSCODE_BUILD_STAMP)
	$(eval CURRENT_ANTLR_GRAMMAR_VERSION := $(shell grep '// Grammar: NeuroScript Version:' $(G4_FILE) | awk '{print $$NF}'))
	@echo ""
	@echo "NeuroScript Build Complete!"
	@echo "--------------------------------------------------"
	@echo "App Version: $(GIT_VERSION), Grammar Version: $(CURRENT_ANTLR_GRAMMAR_VERSION)"
	@echo "Commands installed to: $(BIN_INSTALL_DIR)"
	@echo "Codebase index updated in: $(INDEX_OUTPUT_DIR)"
	@echo "VSCode extension package created in: $(VSCODE_EXT_DIR)"
	@echo "--------------------------------------------------"

# --- Installation Target ---
install: $(CMD_BINS)

# --- Universal Pattern Rule for Go Commands ---
$(BIN_INSTALL_DIR)/%: $(shell find $(CMDS_DIR)/$* -name '*.go') $(ALL_PKG_GO_FILES) $(PROJECT_GO_MOD) $(PROJECT_GO_SUM) $(ANTLR_STAMP_FILE)
	$(eval CURRENT_ANTLR_GRAMMAR_VERSION := $(shell if [ -f "$(G4_FILE)" ]; then grep '// Grammar: NeuroScript Version:' $(G4_FILE) | awk '{print $$NF}'; else echo "g4_not_found"; fi))
	$(eval CURRENT_LDFLAGS := -ldflags="$(LDFLAGS_BASE)$(CURRENT_ANTLR_GRAMMAR_VERSION)")
	@echo "Installing command '$*' to $(BIN_INSTALL_DIR)..."
	$(GO) install $(GOFLAGS) $(CURRENT_LDFLAGS) $(CMDS_DIR)/$*

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

# --- Target to always build the codebase index ---
always-build-index:
	@echo "Building codebase index into $(INDEX_OUTPUT_DIR)..."
	@mkdir -p $(INDEX_OUTPUT_DIR)
	$(GOINDEXER) -root . -output $(INDEX_OUTPUT_DIR)

# --- Target to compile and run the syntax checker ---
syntax-check: $(SYNTAX_CHECKER)
	@echo "Checking syntax of all .ns test files..."
	$(SYNTAX_CHECKER) $(ROOT_DIR)
	
# --- Rules for nslsp & VSCode Extension ---
$(VSCODE_SERVER_DIR)/nslsp_executable: $(BIN_INSTALL_DIR)/nslsp
	@echo "Copying nslsp executable to $(VSCODE_SERVER_DIR)/..."
	@mkdir -p $(VSCODE_SERVER_DIR)
	cp $< $@
	chmod +x $@

$(VSCODE_BUILD_STAMP): $(VSCODE_SERVER_DIR)/nslsp_executable \
						$(shell find $(VSCODE_EXT_DIR) -maxdepth 1 -name 'package.json') \
						$(shell find $(VSCODE_EXT_DIR) \( -name '*.json' -o -name '*.js' -o -name '*.ts' -o -name '*.tsx' -o -name '*.md' \) -type f ! -path '$(VSCODE_EXT_DIR)/server/*' ! -path '$(VSCODE_EXT_DIR)/node_modules/*')
	$(eval CURRENT_GRAMMAR_VERSION := $(shell grep '// Grammar: NeuroScript Version:' $(G4_FILE) | awk '{print $$NF}'))
	@echo "Packaging VSCode extension version $(CURRENT_GRAMMAR_VERSION) in $(VSCODE_EXT_DIR)..."
	cd $(VSCODE_EXT_DIR) && $(VSCE) package $(CURRENT_GRAMMAR_VERSION)
	@touch $@

# --- Convenience Target for Neovim Users ---
setup-nvim: install
	@echo "Setting up NeuroScript plugin for Neovim..."
	$(eval NVIM_PACK_DIR := $(HOME)/.config/nvim/pack/vendor/start)
	@if [ -d "$(VIM_PLUGIN_DIR)" ]; then \
		echo "Plugin source found at $(VIM_PLUGIN_DIR)"; \
		echo "Ensuring Neovim package directory exists at $(NVIM_PACK_DIR)..."; \
		mkdir -p $(NVIM_PACK_DIR); \
		echo "Creating symbolic link for neuroscript plugin..."; \
		ln -s -f $(VIM_PLUGIN_DIR) $(NVIM_PACK_DIR)/neuroscript; \
		echo "Neovim setup complete. Restart Neovim to activate the plugin."; \
	else \
		echo "ERROR: Vim plugin directory not found at $(VIM_PLUGIN_DIR)."; \
		exit 1; \
	fi

# --- Convenience Target for GoLand Users ---
setup-goland: install
	@echo "--------------------------------------------------"
	@echo "GoLand Setup Instructions"
	@echo "--------------------------------------------------"
	@echo "Manual configuration is required for GoLand. Instructions are in GOLAND_SETUP.md."
	@echo "Displaying contents now:"
	@echo ""
	@cat $(GOLAND_SETUP_DOC)

# --- Test Target ---
test: $(ANTLR_STAMP_FILE)
	@echo "Running Go tests for all packages (./...)..."
	$(GO) test ./...
	
# --- Clean Target ---
clean:
	@echo "Cleaning build artifacts..."
	-rm -rf $(ANTLR_OUTPUT_DIR)
	-rm -f $(G4_TXT_FILE)
	-rm -rf $(INDEX_OUTPUT_DIR)
	-rm -f $(VSCODE_BUILD_STAMP)
	-rm -f $(VSCODE_SERVER_DIR)/nslsp_executable
	-rm -f $(VSCODE_EXT_DIR)/*.vsix
	-$(foreach bin,$(CMD_BINS),rm -f $(bin);)
	@echo "Clean complete."

# --- Help Target ---
help:
	@echo "NeuroScript Project Makefile"
	@echo "----------------------------"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Core Targets:"
	@echo "  all          - Checks syntax, then builds and installs all components (default)."
	@echo "  build        - Alias for 'all'."
	@echo "  install      - Compiles and installs Go commands to '$(BIN_INSTALL_DIR)'."
	@echo "  test         - Regenerates parser if needed, then runs Go tests."
	@echo "  clean        - Remove all build artifacts and installed commands."
	@echo "  syntax-check - Checks all .ns files in testdata for syntax errors."
	@echo ""
	@echo "Component & Setup Targets:"
	@echo "  generate-antlr      - Force generation of ANTLR parser files."
	@echo "  always-build-index  - Force regeneration of the codebase index."
	@echo "  setup-nvim          - Link the vim plugin for local Neovim development."
	@echo "  setup-goland        - Display instructions for setting up GoLand."
	@echo "  $(VSCODE_BUILD_STAMP) - Package the VSCode extension."