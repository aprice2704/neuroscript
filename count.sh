#! /bin/bash

cloc \
-exclude-ext=html,js,class,java \
-exclude-lang=JSON \
-exclude-dir=site,tmp,vscode-neuroscript,vim-neuroscript \
*
