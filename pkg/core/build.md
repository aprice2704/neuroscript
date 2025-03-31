# YACC -- deprecated

go install golang.org/x/tools/cmd/goyacc@latest

goyacc -o neuroscript.y.go -p "yy" neuroscript.y

./gonsi/gonsi gonsi/skills HandleSkillRequest "Create a NeuroScript skill that reverses a given input string"

# ANTLR -- current

Install ANTLR and java

In pkg/core do:

java -jar /home/aprice/.m2/repository/org/antlr/antlr4/4.13.2/antlr4-4.13.2-complete.jar  -Dlanguage=Go -o generated -visitor -listener -package core NeuroScript.g4

./gonsi -debug-ast -debug-tokens skills Add 1 2

./gonsi skills HandleSkillRequest "Create a NeuroScript skill that reverses a given input string"