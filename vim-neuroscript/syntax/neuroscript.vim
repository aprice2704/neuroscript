" Vim syntax file for NeuroScript
"
" Language:    NeuroScript
" Maintainer:  Andrew Price (and Gemini)
" Last Change: June 20, 2025

if exists("b:current_syntax")
  finish
endif

" Keywords
syn keyword nsKeyword func means returns needs optional on error do endon event
syn keyword nsKeyword emit must mustbe fail clear_error clear_event ask into
syn keyword nsKeyword if else endif while endwhile for each in endfor
syn keyword nsKeyword break continue call return set
syn keyword nsKeyword and or not no some
syn keyword nsBuiltin ln log sin cos tan asin acos atan len typeof
syn keyword nsConstant true false nil last

" Literals
syn region nsString start=/"/ end=/"/
syn region nsString start=/'/ end=/'/
syn match   nsNumber "\<\d\+\(\.\d\+\)\?\>"

" Comments
syn match nsComment "#.*$"
syn match nsComment "--.*$"
syn match nsComment "//.*$"

" Link the syntax groups to standard highlight groups
hi def link nsKeyword      Keyword
hi def link nsBuiltin      Function
hi def link nsConstant     Constant
hi def link nsString       String
hi def link nsNumber       Number
hi def link nsComment      Comment

let b:current_syntax = "neuroscript"
