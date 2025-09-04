#!/bin/bash

# 1) Coverage (overall + top packages)
go test ./... -coverprofile=coverage.out >/dev/null
echo -n "COVERAGE_TOTAL="; go tool cover -func=coverage.out | awk '/^total:/{print $3}'

# Optional: top 10 packages by uncovered funcs (quick triage)
go tool cover -func=coverage.out \
  | grep -v '^total:' \
  | awk -F'\t' '{sub("%","",$3); if ($3 < 80) print $0}' \
  | sort -t$'\t' -k3,3n | head -10

# 2) Test counts (your gotestsum already prints summary)
echo -n "TEST_FUNCS="; git ls-files '*_test.go' | xargs -I{} grep -h -o 'func Test[A-Za-z0-9_]+' {} | wc -l
# If you use gotestsum, keep its last line (you reported): 
# DONE 1771 tests, 1 skipped, 3 failures in 1.300s

# 3) Cyclomatic complexity (distribution + thresholds)
# Requires gocyclo installed.
gocyclo . | awk '{print $1}' | sort -n > .cyclo.tmp

# Percentiles (min, p50, p90, p99, max)
awk 'NR==1{min=$1} {a[NR]=$1} END{
  n=NR;
  printf "CYCLO_MIN=%s\n", a[1];
  printf "CYCLO_P50=%s\n", a[int(0.50*n)];
  printf "CYCLO_P90=%s\n", a[int(0.90*n)];
  printf "CYCLO_P99=%s\n", a[int(0.99*n)];
  printf "CYCLO_MAX=%s\n", a[n];
}' .cyclo.tmp

# Over-threshold counts (classic guardrails)
echo -n "CYCLO_>15="; gocyclo -over 15 . | wc -l
echo -n "CYCLO_>25="; gocyclo -over 25 . | wc -l

# 4) Build health (quick signal)
echo -n "GO_BUILD_CLEAN="; (go build . >/dev/null 2>&1 && echo yes) || echo no
