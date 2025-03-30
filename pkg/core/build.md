

go install golang.org/x/tools/cmd/goyacc@latest

goyacc -o neuroscript.y.go -p "yy" neuroscript.y

./gonsi/gonsi gonsi/skills HandleSkillRequest "Create a NeuroScript skill that reverses a given input string"