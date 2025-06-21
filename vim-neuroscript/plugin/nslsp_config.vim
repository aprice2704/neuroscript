" Vim plugin for NeuroScript LSP Configuration
"
" Author: Andrew Price (and Gemini)
" Last Change: June 20, 2025

if !executable('nslsp')
  echom "nslsp not found. Please run 'make install' in the neuroscript repository and ensure the binary is in your PATH."
  finish
endif

" --- Neovim lspconfig setup ---
if has('nvim') && luaeval('pcall(require, "lspconfig")')
  lua << EOF
    require('lspconfig').nslsp.setup{}
EOF
  finish
endif

" --- vim-lsp setup ---
if exists(':LspRegisterServer')
  augroup lsp_neuroscript
    autocmd!
    autocmd User lsp_setup call lsp#register_server({
          \ 'name': 'nslsp',
          \ 'cmd': {server_info->['nslsp']},
          \ 'root_uri': {server_info->lsp#utils#get_buffer_root_uri(server_info.bufnr)},
          \ 'whitelist': ['neuroscript'],
          \ })
  augroup END
  finish
endif
