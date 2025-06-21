" Vim filetype detection for NeuroScript
"
" Author: Andrew Price (and Gemini)
" Last Change: June 20, 2025

" Detect .ns or .ns.txt files and set their filetype to 'neuroscript'
au BufRead,BufNewFile *.ns,*.ns.txt setfiletype neuroscript
