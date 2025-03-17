package main

import (
	"fmt"
	"os"
	"github.com/pborman/getopt/v2" // version 2
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/michalmaruska/git-hierarchy/git_hierarchy"
	"github.com/michalmaruska/git-hierarchy/graph"
)

func usage() {
	getopt.PrintUsage(os.Stderr)
}

func main() {
	helpFlag := getopt.BoolLong("help", 'h', "display help")
	// no errors, just fail:
	getopt.SetUsage(func() {
		getopt.PrintUsage(os.Stderr)
		fmt.Println("\nparameter:  from  to")
	})
	getopt.Parse() // os.Args

	if *helpFlag {
		// I want it to stdout!
		fmt.Println(plumbing.RefRevParseRules)
		getopt.Usage()
		os.Exit(0)
	}

	// plan: collect the graph, linearized,
	// current_branch()
	repository, err := git.PlainOpen(".")
	git_hierarchy.TheRepository = repository

	current, err := repository.Reference(plumbing.ReferenceName("HEAD"), true)
	fmt.Println("Current head is", current.Name())
	git_hierarchy.CheckIfError(err, "finding repository")

	git_hierarchy.WalkHierarchy(current)
	fmt.Println(current)
	// walk_down_from()
	// mark
	// unmark

	// state
	// gitRebasePoset()

	os.Exit(0)
}
