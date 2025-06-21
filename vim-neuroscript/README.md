# NeuroScript for Vim

Provides syntax highlighting and Language Server Protocol (LSP) support for NeuroScript (.ns, .ns.txt) files in Vim and Neovim.

This plugin automatically configures itself to use nslsp, the official language server for NeuroScript.



## Features

* Syntax Highlighting: For all keywords, literals, and comments.
* LSP Integration (via nslsp):
    * Live diagnostics (errors and warnings)
    * Code completion (future nslsp feature)
    * Go to definition (future nslsp feature)
    * Hover information (future nslsp feature)

## Requirements

1.  nslsp Executable: The plugin requires the nslsp language server executable to be in your system's $PATH. You can build and install it from the root of the NeuroScript repository:
    sh make install 

2.  A Vim/Neovim LSP Client: This plugin doesn't include an LSP client, it only configures one. You need one of the following:
    * Neovim: nvim-lspconfig (standard for native LSP).
    * Vim 8+: vim-lsp or coc.nvim.

## Installation

Install using your favorite Vim plugin manager.

### vim-plug

vim Plug 'aprice2704/neuroscript', {'rtp': 'vim-neuroscript'} 

### packer.nvim (for Neovim)

lua use { 'aprice2704/neuroscript', rtp = 'vim-neuroscript' } 

### Vim 8 Native Packages

sh # Create the necessary directory mkdir -p ~/.vim/pack/vendor/start # Clone the repository git clone https://github.com/aprice2704/neuroscript.git ~/.vim/pack/vendor/start/neuroscript 
Note: Because the plugin is in a subdirectory of the main repo, you need to add it to Vim's runtime path. For native packages, this is handled automatically if you clone the whole repo. For other managers, use the rtp (runtime path) option as shown above.

## Configuration

The plugin attempts to automatically detect your LSP client (nvim-lspconfig or vim-lsp) and configure the nslsp server. No further configuration should be necessary.