package main

import (
	"fmt"
	"log"
	"os"
	"github.com/michalmaruska/git-hierarchy/git_hierarchy"
	"github.com/michalmaruska/git-hierarchy/graph"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pborman/getopt/v2"
)



func cloneHierarchy(vertices *[]graph.NodeExpander, order []int, prefix string) {
	// assignment to entry in nil map
	var renamed  = make(map[string]*plumbing.Reference) // Name
	// so what's the difference between this & make() ?

	// can I simplify this iterator?
	for i := range order {
		gh := git_hierarchy.GetHierarchy((*vertices)[i])

		fmt.Println("cloning", gh.Name() )
		// skip over?

		//
		switch gh.(type) {
		case git_hierarchy.Segment:
			segment := gh.(git_hierarchy.Segment)

			name := segment.Name()
			// ignore?
			newName :=prefix + name

			// newRef := plumbing.ReferenceName(git_hierarchy.HeadPrefix + newName)
			fmt.Println("new segment ", newName)

			base := segment.Base
			if val, found := renamed[base.Name().String()]; found {
				// so what is
				base = val
			}

			newSegment := git_hierarchy.MakeSegment(
				newName, base,
				segment.Ref.Hash(),
				segment.Start.Hash())
			/*
			hash := segment.Ref.Hash()
			newSegment := git_hierarchy.Segment{
				Ref: plumbing.NewHashReference(newName, hash),
				Base: segmentBase(newName),
				Start: segment.Start}
			*/
			fmt.Println("Segment", newSegment.Ref.Name(),
				"base", newSegment.Base.Name(),
				"->", newSegment.Base.Target(),
				"start", newSegment.Start.Name(), "->", hash)
			newSegment.Write()

			// how to create it?
			renamed[name] = newSegment.Ref

		case git_hierarchy.Sum:
			sum := gh.(git_hierarchy.Sum)
			name := sum.Name()
			newName := plumbing.ReferenceName(prefix + name)

			hash := sum.Ref.Hash()
			newSum := git_hierarchy.Sum{
				Ref: plumbing.NewHashReference(newName, hash),
				Summands: make([]*plumbing.Reference, len(sum.Summands)) }

			// list-of new summands.

			// create sum, with summands....
			for i, summand := range sum.Summands {
				newTarget := summand.Target()
				// points into renamed[]
				// then rewrite.

				if val, found := renamed[newTarget.String()]; found {
					newTarget = val.Name()
				}
				newSum.Summands[i] = plumbing.NewSymbolicReference(
					summand.Name(),
					newTarget)
			}
			// for summand ... renamed ?

		default:
		}
	}
}



//  mixed named and unnamed parameters
func replaceInHierarchy(vertices *[]graph.NodeExpander, order []int, from string, replacement string) {
	// todo: replaceFlag must be a ReferenceName -- existing!
	// skip also!
	repository := git_hierarchy.TheRepository

	_, err := repository.Reference(plumbing.ReferenceName(from), false)
	git_hierarchy.CheckIfError(err, "the replacement match is invalid")

	ref2, err := repository.Reference(plumbing.ReferenceName(replacement), false)
	git_hierarchy.CheckIfError(err, "the replacement is invalid")


	for i := range order {
		// act:
		gh := git_hierarchy.GetHierarchy((*vertices)[i])

		switch gh.(type) {
		case git_hierarchy.Segment:
			segment := gh.(git_hierarchy.Segment)
			if segment.Base.Target().String() == from {
				println("let's replace in segment", segment.Name())
				println(segment.Base.Target(), "vs", from)

				segment.SetBase(ref2)
			}
		case git_hierarchy.Sum:
			//
		default:
		}
	}
}


func dump(vertices *[]graph.NodeExpander, order []int) {
	for _, j := range order {
		a := (*vertices)[j]
		gh := git_hierarchy.GetHierarchy(a)

		switch v := gh.(type) {
		case  git_hierarchy.Segment:
			fmt.Println("segment", v.Name())
		case  git_hierarchy.Sum:
			fmt.Println("sum", v.Name())
		case  git_hierarchy.Base:
			fmt.Println("plain base reference", v.Name())
		default:
			fmt.Println("unexpected git_hierarchy type")
			// error("unexpected")
		}
	}
}

func main() {

	// os.Args[0] == "git-walk-down"
	set := getopt.New()
	helpFlag := set.BoolLong("help", 'h', "display help")


	skipOpt := set.StringLong("skip", 's', "", "skip")
	replaceFlag := set.StringLong("replace", 't', "", "replace")

	cloneOpt := set.StringLong("clone", 'c', "", "clone using a prefix")



	// cloneFlag := getopt.BoolLong("clone", 'c', "clone hierarchy, prefix")

	// no errors, just fail:
	set.SetUsage(func() {
		getopt.PrintUsage(os.Stderr)
		fmt.Println("\nparameter:  from  to")
	})

	// var opts = getopt.CommandLine
	set.Parse(os.Args)
	if *helpFlag {
		// I want it to stdout!
		fmt.Println(plumbing.RefRevParseRules)
		getopt.Usage()
		os.Exit(0)
	}

	// sanity check:
	if (*replaceFlag != "") != (*skipOpt != "") {
		log.Fatal("replace & match must come in pair!")
		getopt.Usage()
	}


	repository, err := git.PlainOpen(".")
	git_hierarchy.TheRepository = repository


	// ---------------------------
	args := set.Args()
	var top *plumbing.Reference

	// var err Error
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
	// ---------------------------

	vertices, incidenceGraph := git_hierarchy.WalkHierarchy(top)
	if false {
		fmt.Println("Visited these git refs:")
		for i, v := range *vertices {
			fmt.Println(i, "->", v.NodeIdentity())
		}

		fmt.Println("Now edges:")
		graph.DumpGraph(incidenceGraph)
	}

	order, err := graph.TopoSort(incidenceGraph)
	git_hierarchy.CheckIfError(err)

	if *replaceFlag != "" && *skipOpt != "" {
		replaceInHierarchy(vertices, order, *skipOpt, *replaceFlag)
	}

	if *cloneOpt != "" {
		cloneHierarchy(vertices, order, *cloneOpt)
	}

	// dump index -> vertices[i]
	// fmt.Println("order:", order)

	dump(vertices, order)
	os.Exit(0)
}
