#!/bin/zsh -feu

set -x
go vet ./...

go vet -vettool=/home/michal/go/bin/shadow ./...
# I built it myself:
# /home/michal/go/pkg/mod/golang.org/x/tools@v0.31.0/README.md

# go install golang.org/x/tools/cmd/goimports@latest
# go build /home/michal/go/pkg/mod/golang.org/x/tools@v0.31.0/go/analysis/passes/shadow/cmd/shadow/main.go
# mv main ~/go/bin/shadow

# /home/michal/go/pkg/mod/golang.org/x/tools@v0.31.0/go/analysis/passes/shadow/cmd/shadow ./...

~/go/bin/ineffassign ./...

# -v print the names of packages as they are compiled
go install -v ./...

