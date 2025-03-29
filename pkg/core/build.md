# Building the lexer

go install golang.org/x/tools/cmd/goyacc@latest

goyacc -o neuroscript.y.go -p "yy" neuroscript.y