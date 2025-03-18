#!/bin/zsh -feu

set -x
go vet ./...

# -v print the names of packages as they are compiled
go install -v ./...
