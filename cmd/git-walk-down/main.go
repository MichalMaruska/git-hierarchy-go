package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/michalmaruska/git-hierarchy/git_hierarchy"
	"github.com/michalmaruska/git-hierarchy/graph"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pborman/getopt/v2"
)

var verbose = false

func remap(ref plumbing.ReferenceName,
	remapped map[string]*plumbing.Reference) (bool, plumbing.ReferenceName) {
	if val, found := remapped[ref.String()]; found {
		// fmt.Println("rewriting")
		ref = val.Name()
		return true, ref
	}
	return false, ref
}

func cloneSegment(segment git_hierarchy.Segment, newName string,
	remapped map[string]*plumbing.Reference) git_hierarchy.Segment {

	if verbose {
		fmt.Println("new segment", newName)
	}
	_, base := remap(segment.Base.Target(), remapped)

	newSegment := git_hierarchy.MakeSegment(
		newName,
		base,
		segment.Ref.Hash(),
		segment.Start.Hash())
	if verbose {
		fmt.Println("Segment", newSegment.Ref.Name(), // hash
			"base", newSegment.Base.Name(),
			"->", newSegment.Base.Target(),
			"start", newSegment.Start.Name(), "->", newSegment.Start.Hash())
	}
	newSegment.Write()
	return newSegment
}

func cloneSum(sum git_hierarchy.Sum, newName string,
	remapped map[string]*plumbing.Reference) git_hierarchy.Sum {

	commitId := sum.Ref.Hash()
	newSum := git_hierarchy.MakeSum(newName, commitId,
		make([]*plumbing.Reference, len(sum.Summands)))

	// list-of new summands.
	if verbose {
		fmt.Println("Look at summands")
	}

	// convert the summands....
	for i, summand := range sum.Summands {
		_, newTarget := remap(summand.Target(), remapped)

		number := git_hierarchy.SumSummandIndex(sum.Name(), summand.Name())

		if verbose {
			fmt.Println("another summand", number, newTarget)
		}

		newSummandName := git_hierarchy.SumSummand(newName, number)

		newSum.Summands[i] = plumbing.NewSymbolicReference(
			newSummandName,
			newTarget)
	}

	// for summand ... remapped ?
	newSum.Write()
	return newSum
}

func cloneHierarchy(vertices *[]graph.NodeExpander, order []int, prefix string,
	renamed map[string]*plumbing.Reference) {
	// assignment to entry in nil map
	// so what's the difference between this & make() ?

	// can I simplify this iterator?
	for i := range slices.Backward(order) {
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
			newSegment := cloneSegment(segment, newName, renamed) // maps

			// how to create it?
			// or renamed[] ?
			renamed[segment.Ref.Name().String()] = newSegment.Ref

		case git_hierarchy.Sum:
			sum := gh.(git_hierarchy.Sum)
			name := sum.Name()
			newName := prefix + name
			newSum := cloneSum(sum, newName, renamed)
			renamed[sum.Ref.Name().String()] = newSum.Ref
		default:
		}
	}
}



//  mixed named and unnamed parameters
func replaceInHierarchy(vertices *[]graph.NodeExpander, order []int, remapped map[string]*plumbing.Reference) {

	// todo: replaceFlag must be a ReferenceName -- existing!
	// skip also!
	for i := range order {
		// act:
		gh := git_hierarchy.GetHierarchy((*vertices)[i])

		switch gh.(type) {
		case git_hierarchy.Segment:
			segment := gh.(git_hierarchy.Segment)

			found, value := remap(segment.Base.Target(), remapped)
			if found {
				println("let's replace in segment", segment.Name())
				println(segment.Base.Target(), "vs", value)

				segment.SetBase(value)
			}
		case git_hierarchy.Sum:
			//
		default:
		}
	}
}


func dump(vertices *[]graph.NodeExpander, incidenceGraph *graph.Graph, order []int) {
	if false {
		fmt.Println("Visited these git refs:")
		for i, v := range *vertices {
			fmt.Println(i, "->", v.NodeIdentity())
		}

		fmt.Println("Now edges:")
		graph.DumpGraph(incidenceGraph)
	}

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
	// helpFlag
	_ = set.BoolLong("help", 'h', "display help")

	skipOpt := set.StringLong("skip", 's', "", "skip reference")
	replaceFlag := set.StringLong("replace", 't', "", "and replace with another reference")

	cloneOpt := set.StringLong("clone", 'c', "", "clone using a prefix")

	set.SetParameters("[top-reference]")

	// I don't use Parse(), but Getopt() so!!!
	// set.SetUsage(func() {usage(set)}) // to display usage on error.
	// var opts = getopt.CommandLine

	var current = ""
	var remapped = make(map[string]*plumbing.Reference)

	repository, err := git_hierarchy.FindGitRepository()
	git_hierarchy.CheckIfError(err, "looking for .git repo")
	git_hierarchy.TheRepository = repository

	if err = set.Getopt(os.Args, func(o getopt.Option) bool {
		// fmt.Println("looking at option: ", o.LongName())
		switch o.LongName() {
		case "help":
			set.PrintUsage(os.Stdout)
			os.Exit(0)
		case "skip": {
			current = o.Value().String()
		}
		case "replace": {
			from := current
			replacement := o.Value().String()

			fmt.Println("let's resolve the values: ", from, "and", replacement)
			_, err = repository.Reference(plumbing.ReferenceName(from), false)
			git_hierarchy.CheckIfError(err, "the replacement match is invalid")

			var ref2 *plumbing.Reference
			ref2, err = repository.Reference(plumbing.ReferenceName(replacement), false)
			git_hierarchy.CheckIfError(err, "the replacement is invalid")

			log.Print("Will replace any use of ", from, "with reference to ", ref2)
			remapped[from] = ref2
		}
		}
		return true }); err != nil {

		fmt.Fprintln(os.Stderr, err)
		set.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	// sanity check:
	if (*replaceFlag != "") != (*skipOpt != "") {
		log.Fatal("replace & match must come in pair!")
		set.PrintUsage(os.Stderr)
		os.Exit(2)
	}

	args := set.Args()
	var top *plumbing.Reference
	if len(args) > 0 {
		top = git_hierarchy.FullHeadName(repository, args[0])
		if top == nil {
			os.Exit(-1)
		}
		fmt.Println("Will descend from", top.Name())
	} else {
		var currentBranch *plumbing.Reference
		currentBranch, err = repository.Head()
		git_hierarchy.CheckIfError(err)
		fmt.Println("Current head is", currentBranch.Name())
		top = currentBranch
	}

	vertices, incidenceGraph := git_hierarchy.WalkHierarchy(top)

	order, err := graph.TopoSort(incidenceGraph)
	git_hierarchy.CheckIfError(err)


	if *cloneOpt != "" {
		cloneHierarchy(vertices, order, *cloneOpt, remapped)
	} else if *replaceFlag != "" && *skipOpt != "" {
		replaceInHierarchy(vertices, order, remapped)
	}


	dump(vertices, incidenceGraph, order)
	os.Exit(0)
}
