# Rebuild grammar

## Current

goyacc -o checklist_parser.go -p "yy" checklist.y


## Old

java -jar ../../antlr4-4.13.2-complete.jar  -Dlanguage=Go -o generated -visitor -listener -package generated NeuroDataChecklist.g4