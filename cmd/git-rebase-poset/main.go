package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/pborman/getopt/v2" // version 2
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/michalmaruska/git-hierarchy/git_hierarchy"
	"github.com/michalmaruska/git-hierarchy/graph"
)

func usage() {
	getopt.PrintUsage(os.Stderr)
}

var verbose = false
func startRebase(top *plumbing.Reference, noFetch bool) {
	vertices, incidenceGraph := git_hierarchy.WalkHierarchy(top)

	order, err := graph.TopoSort(incidenceGraph)
	git_hierarchy.CheckIfError(err)

	/* todo:
	NewReferenceSliceIter(
	    Map(order, func (index int) { git_hierarchy.GetHierarchy((*vertices)[index]) }
	*/
	for _, j := range slices.Backward(order) {
		gh := git_hierarchy.GetHierarchy((*vertices)[j])

		if verbose {
			fmt.Println("\n** Processing:", gh.Name())
		}
		switch gh.(type) {
		case git_hierarchy.Base:
			if !noFetch {
				git_hierarchy.FetchUpstreamOf(gh.(git_hierarchy.Base).Ref)
			}
		default:
			status := git_hierarchy.RebaseNode(gh)
			if status == git_hierarchy.RebaseFailed {
				fmt.Println("failed on", gh.Name())
				os.Exit(1)
			}
		}
	}
}

func contRebase(repository *git.Repository) {

	ref, err := repository.Reference(".segment-cherry-pick", false)
	git_hierarchy.CheckIfError(err, "finding mark")

	ref1, err :=  repository.Reference(ref.Target(), false)
	git_hierarchy.CheckIfError(err, "finding segment being rebased")

	// git cherry-pick --abort
	// if cherry-pick finished
	gh := git_hierarchy.Convert(ref1)

	// var result rebaseResult
	switch v := gh.(type) {
	case  git_hierarchy.Segment:
		git_hierarchy.RebaseSegmentFinish(gh.(git_hierarchy.Segment))
	default:
		fmt.Println("unexpected git_hierarchy type", v)
		// error("unexpected")
		// result = RebaseFailed
	}

	// mark := plumbing.NewSymbolicReference(".segment-cherry-pick", segment.Ref.Name())
	err = repository.Storer.RemoveReference(ref.Name())
	git_hierarchy.CheckIfError(err, "removing Mark")

	// echo -n "was rebasing $segment"
	// gitRun("switch", segment)
}

func main() {
	helpFlag := getopt.BoolLong("help", 'h', "display help")
	contFlag := getopt.BoolLong("continue", 'c', "continue after manual fix")
	fetchFlag := getopt.BoolLong("nofetch", 'f', "don't fetch from remote branches")

	// no errors, just fail:
	getopt.SetUsage(func() {
		// the default:
		getopt.PrintUsage(os.Stderr)
	})
	getopt.SetParameters("[top-reference]")

	getopt.Parse() // os.Args

	if *helpFlag {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	// plan: collect the graph, linearized,
	repository, err := git_hierarchy.FindGitRepository()
	git_hierarchy.CheckIfError(err, "finding repository")
	git_hierarchy.TheRepository = repository


	if *contFlag {
		contRebase(repository)
		os.Exit(0)
	}

	args := getopt.Args()

	var top *plumbing.Reference

	if len(args) > 0 {
		// := leads to crash!
		top = git_hierarchy.FullHeadName(repository, args[0])
		if top == nil {
			os.Exit(-1)
		}
		fmt.Println("Will descend from", top.Name())
	} else {
		current, err := repository.Head()
		git_hierarchy.CheckIfError(err)
		fmt.Println("Current head is", current.Name())
		top = current
	}

	startRebase(top, *fetchFlag)

	os.Exit(0)
}
