# Rebuild grammar

java -jar ../../antlr4-4.13.2-complete.jar  -Dlanguage=Go -o generated -visitor -listener -package generated NeuroDataChecklist.g4