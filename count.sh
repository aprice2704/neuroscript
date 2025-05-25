#! /bin/bash

cloc \
-exclude-ext=html,js \
-exclude-lang=JSON \
-exclude-dir=site  \
-exclude-dir=vscode-neuroscript \
*
