# YACC -- deprecated

go install golang.org/x/tools/cmd/goyacc@latest

goyacc -o neuroscript.y.go -p "yy" neuroscript.y

./neurogo/neurogo neurogo/skills HandleSkillRequest "Create a NeuroScript skill that reverses a given input string"

# ANTLR -- current

Install ANTLR and java

In pkg/core do:

java -jar antlr4-4.13.2-complete.jar  -Dlanguage=Go -o generated -visitor -listener -package core NeuroScript.g4

./neurogo -debug-ast -debug-tokens skills Add 1 2

./neurogo skills HandleSkillRequest "Create a NeuroScript skill that reverses a given input string"