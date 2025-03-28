package git_hierarchy

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
	"os"
	// "io"
	"github.com/go-git/go-git/v5"
	// how come this git, not go-git ? or is that the whole module with git package inside?

	// . "github.com/go-git/go-git/v5/_examples"
	"github.com/go-git/go-git/v5/plumbing"
	_ "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/samber/lo"
	lom "github.com/samber/lo/mutable"
	// "github.com/kendru/darwin/go/depgraph"
	"github.com/michalmaruska/git-hierarchy/graph"
)


// todo: move into common file:
func CheckIfError(err error, msgs ...string) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))

	if len(msgs) > 0 {
		for _, msg := range msgs {
			fmt.Println(msg)
		}
	}

	os.Exit(1)
}


func Unimplemented(){
	fmt.Fprintln(os.Stderr, "Not implemented")
	os.Exit(1)
}

func FindGitRepository() (*git.Repository, error){
	openOptions := git.PlainOpenOptions{
		DetectDotGit: true,
	}
	err := openOptions.Validate()
	CheckIfError(err, "options")
	return git.PlainOpenWithOptions(".", &openOptions)
}


// MatchAny returns true if any of the RefSpec match with the given ReferenceName.
// Massage the user-provided branch name into full reference.
func FullHeadName(repository *git.Repository, refName string) *plumbing.Reference {
	// HEAD -> ?
	// name
	// heads/name
	for _,pattern := range plumbing.RefRevParseRules {
		s := fmt.Sprintf(pattern, refName)
		// existence test:
		// repository.storer.Reference(plumbing.ReferenceName) (*plumbing.Reference, error)
		ref2, error := repository.Reference(plumbing.ReferenceName(s), false) // not resolved
		if (error == nil) {
			// fmt.Println("found", ref2.Name(), s)
			return ref2
		}
	}
	return nil
}

// const
var HeadPrefix string

// func (s, suffix string) (before string, found bool)
// "refs/heads/"
func init(){
	// note:
	HeadPrefix, _ = strings.CutSuffix(plumbing.RefRevParseRules[3], "%s")
}

const sumPattern = "refs/sums/%s"
const sumSummandPattern = "refs/sums/%s/%d"
const sumSummandPrefix = "refs/sums/"
const segmentBasePattern = "refs/base/%s"
const segmentStartPattern = "refs/start/%s"


// in the storer?
func referenceExists(repository *git.Repository, name string) bool {
	// not resolved
	_, error := repository.Reference(plumbing.ReferenceName(name), false)
	return (error == nil)
}

func SumSummand(name string, n int) plumbing.ReferenceName {
	return plumbing.ReferenceName(fmt.Sprintf(sumSummandPattern, name, n))
}

func SumSummandIndex(sumname string, summand plumbing.ReferenceName) int {

	re := regexp.MustCompile(`refs/sums/(.*)/([[:digit:]]*)`)
	matches := re.FindStringSubmatch(summand.String())

	// assert sumname == sum.Name()
	if (sumname != matches[1]) {
		panic("bad")
	}

	i, _ := strconv.Atoi(matches[2])
	return i
}


/* todo: */

func refsWithPrefixIter(iterator storer.ReferenceIter, prefix string) storer.ReferenceIter {
	return storer.NewReferenceFilteredIter (
		func (ref *plumbing.Reference) bool {
			return strings.HasPrefix(ref.Name().String(), prefix)},
		iterator)
}

func refsWithPrefix(repository *git.Repository, prefix string) []*plumbing.Reference {
	collector := []*plumbing.Reference{}

	refIter, _ := repository.References()
	// fmt.Fprintln(os.Stderr, "looking for this prefix:", prefix)
	refIter.ForEach( func(ref *plumbing.Reference) error {
		// fmt.Fprintln(os.Stderr, "looking at", ref.Name().String())

		if strings.HasPrefix(ref.Name().String(), prefix) {
			// fmt.Fprintln(os.Stderr,"found")
			collector = append(collector, ref) //  yield
			// found = branch
			// return ErrStop
		}

		return nil
	})

	// fmt.Fprintln(os.Stderr,"returning", len(collector))
	return collector
}

//  { ref; ref ~prefix &&  which expand to  "ref: prefix ...." }
// given a prefix, find all refs, whose name matches?
// todo: make a goroutine: filter, and search the contents.
func symbolic_refs_to(repository *git.Repository, ref *plumbing.Reference, prefix string) []*plumbing.Reference {
	// collector := refsWithPrefix(repository, prefix)
	refIter, _ := repository.References()
	iter := refsWithPrefixIter(refIter, prefix)

	var refs []*plumbing.Reference
	// todo: a function for this:
	s := "ref: " + ref.Name().String()

	// for _, ref := range collector {
	iter.ForEach ( func(ref *plumbing.Reference) error {
		content := dump_symbolic_ref(ref)
		// reduced, _ := strings.CutPrefix(content, "ref: ")
		if s == content.String() {
			refs = append(refs, ref)
		}
		return nil})

	return refs
}

/// Not exported!

//   { segment, @ref is base of }
func base_of(repository *git.Repository, ref *plumbing.Reference)  []*plumbing.Reference {
	// iterate over given prefix
	return symbolic_refs_to(repository, ref, "refs/base/")
}

//  { sum,  @ref is summand of sum}
func summand_of(repository *git.Repository, ref *plumbing.Reference)  []*plumbing.Reference{
	return symbolic_refs_to(repository, ref, "refs/sums/")
}

// return collection
func sumSummands(repository *git.Repository, name string) []*plumbing.Reference {
	return refsWithPrefix(repository,  sumSummandPrefix + name + "/")
}

// @name is always without refs/heads
func isSum(name string, repository *git.Repository) (bool, []*plumbing.Reference) {
	// expand all subtree
	summands := sumSummands(repository, name)
	return len(summands) > 0, summands
}

func segmentBase(name string) plumbing.ReferenceName {
	return plumbing.ReferenceName(fmt.Sprintf(segmentBasePattern, name))
}

func segmentStart(name string) plumbing.ReferenceName {
	return plumbing.ReferenceName(fmt.Sprintf(segmentStartPattern, name))
}

func isSegment(name string, repository *git.Repository) (bool, *plumbing.Reference) {
	s := segmentBase(name)

	base , err := repository.Reference(s, false)
	return (err == nil), base
}

func dump_symbolic_ref(ref *plumbing.Reference) plumbing.ReferenceName {
	content := ref.Target() // Strings()
	// fmt.Println("symbolic ref", content[0], "points at", content[1])
	return content
}

func set_symbolic_reference(repository *git.Repository,
	refName plumbing.ReferenceName, content plumbing.ReferenceName) *plumbing.Reference {
	// todo: NewReferenceFromStrings(content)
	reduced, _ := strings.CutPrefix(content.String(), "ref: ")

	ref := plumbing.NewSymbolicReference(refName, plumbing.ReferenceName(reduced))
	// ReferenceStorer
	repository.Storer.SetReference(ref)
	return ref
}


func drop_symbolic_ref(repository *git.Repository, ref *plumbing.Reference ) error {
	// git update-ref --no-deref -d $ref
	return repository.Storer.RemoveReference(ref.Name())
}


// given a reference, change its name? the contents remains the same
func rename_symbolic_reference(repository *git.Repository,
	ref *plumbing.Reference, newName plumbing.ReferenceName){

	fmt.Println("rename symbolic:", ref.Name(), "as", newName)

	if ref.Name() == newName {
		return
	}
	var content = dump_symbolic_ref(ref)
	fmt.Println("contains:", content)

	set_symbolic_reference(repository, newName, content)

	err := drop_symbolic_ref(repository, ref)
	CheckIfError(err)
}

var verbose = false

// not symbolic
func rename_reference(repository *git.Repository, ref *plumbing.Reference, newName plumbing.ReferenceName) {

	if verbose {fmt.Println("Renaming ref", ref.Name().String(), "to", newName.String())}

	var content = ref.Hash() // dump_symbolic_ref(ref)
	newRef := plumbing.NewHashReference(newName, content)
	repository.Storer.SetReference(newRef)

	// drop_symbolic_ref(repository, ref)
	err := repository.Storer.RemoveReference(ref.Name())
	CheckIfError(err)
}

// no way to override
func branchName(full plumbing.ReferenceName) string {
	// fmt.Println("extract name from", full.Name().String())
	sName, _ := strings.CutPrefix(full.String(), HeadPrefix) // todo: strip
	return sName
}

/// Public api:
// Given a head (sum or segment), rename it.
func Rename(repository *git.Repository, from string, to string) {
	// var found bool
	full := FullHeadName(repository, from)
	if (full == nil) {
		// return error("the branch does not exist")
		return
	}
	sName := branchName(full.Name())

	// no.... just drop the
	toFull := plumbing.ReferenceName(HeadPrefix + to)
		// FullHeadName(repository, to)


	fmt.Println("would use ", sName, "as base for derived references")

	newName, _ := strings.CutPrefix(to, HeadPrefix)

	// begin transaction!
	rename_reference(repository, full, plumbing.ReferenceName(toFull))

	// todo: duplicate all
	// redirect
	// drop

	// walk all bases, sums,
	// change those:
	above := base_of(repository, full)
	sums_with := summand_of(repository, full)

	var ref *plumbing.Reference

	for _, ref = range above {
		set_symbolic_reference(repository, ref.Name(), toFull)
	}
	for _, ref = range sums_with {
		set_symbolic_reference(repository, ref.Name(), toFull)
	}

	// unimplemented()

	if is, summands := isSum(sName, repository); is {
		fmt.Println("Sum summands")
		// for ...
		// walkReferencesTree
		prefix := sumSummandPrefix + sName + "/"

		for _, s := range summands {
			end, _ := strings.CutPrefix(s.Name().String(), prefix)
			n, _ := strconv.Atoi(end)
			rename_symbolic_reference(repository, s, SumSummand(newName, n))
		}
	}

	if is, base := isSegment(sName, repository); is {
		// fmt.Println("it's a segment!")

		// ref, err := repository.Reference(segmentBase(sName), false)
		// CheckIfError(err)
		rename_symbolic_reference(repository, base, segmentBase(newName))

		ref, err := repository.Reference(segmentStart(sName), false)
		CheckIfError(err)
		rename_reference(repository, ref, segmentStart(newName))
	}

}

var TheRepository *git.Repository

/*
type Rebasable interface{
	rebase
}
*/

type GitHierarchy interface{
	// segment | sum
	Children() []*plumbing.Reference // GitHierarchy
	Name() string
}

type  Segment struct {
	Ref *plumbing.Reference
	Base *plumbing.Reference
	Start *plumbing.Reference // Hash
}

func MakeSegment(name string, base plumbing.ReferenceName, head plumbing.Hash, hash plumbing.Hash) Segment {
	return Segment{
		Ref: plumbing.NewHashReference(plumbing.ReferenceName(HeadPrefix + name), head),
		Base: plumbing.NewSymbolicReference(segmentBase(name), base),
		Start: plumbing.NewHashReference(segmentStart(name), hash),
	}
}

func (s Segment) Write() {
	err := TheRepository.Storer.SetReference(s.Ref)
	CheckIfError(err)
	err = TheRepository.Storer.SetReference(s.Base)
	CheckIfError(err)
	err = TheRepository.Storer.SetReference(s.Start)
	CheckIfError(err)
}


func (s Segment) Name() string {
	return branchName(s.Ref.Name())
}

func setReferenceTo(repository *git.Repository, reference *plumbing.Reference, target *plumbing.Reference) {
	// target: ref, hash, ...
	ref, err := storer.ResolveReference(repository.Storer, target.Name())
	CheckIfError(err)

	hash := ref.Hash() // assert it's a hash?

	// , err := repository.Reference(segment.Base, true)
	// plumbing.NewHash()
	fmt.Println("setting", reference.Name(), "to", target.Name(), "= hash", hash)

	// segment.SetStart(hash)
	newRef := plumbing.NewHashReference(reference.Name(), hash)
	err = repository.Storer.CheckAndSetReference(newRef, reference)
	CheckIfError(err)
}



// set start to base/ref
func (segment Segment) ResetStart() {
	// 1. why symbolic?
	// how to reset to
	// err := repository.Storer.SetReference(*plumbing.Reference)
	// new base:
	// ref, err :=
	// CheckIfError(err)
	setReferenceTo(TheRepository, segment.Start, segment.Base)
}

func (segment Segment) SetBase(replacement plumbing.ReferenceName) {
	set_symbolic_reference(TheRepository, segment.Base.Name(), replacement)
}

// only references
func (s Segment) Children() []*plumbing.Reference {
	repository := TheRepository

	// fmt.Println("Children of", s.Name())
	baseName := segmentBase(s.Name())
	base, err := repository.Reference(plumbing.ReferenceName(baseName), true)

	CheckIfError(err, "while searching for ", baseName.String())
	return []*plumbing.Reference{base}
}


type  Sum struct {
	// why pointer? and it should be const.
	Ref *plumbing.Reference
	Summands []*plumbing.Reference // can be map[int]*plumbing.Reference, or just array
}

func MakeSum(name string, commitId plumbing.Hash, summands []*plumbing.Reference) Sum {
	return Sum{
		Ref: plumbing.NewHashReference(plumbing.ReferenceName(HeadPrefix + name), commitId),
		Summands: summands}
}


// todo: identical. Generics?
func (s Sum) Name() string {
	return branchName(s.Ref.Name())
}

func (s Sum) Children() []*plumbing.Reference {
	repository := TheRepository
	sr := s.Summands
	lom.Reverse(sr)
	return lo.Map(sr,
		func(x *plumbing.Reference, index int) *plumbing.Reference {
			ref, err := repository.Reference(x.Name(), true) // only 1 step!
			CheckIfError(err)
			return ref
		})
}
func (s Sum) Write() {
	err:= TheRepository.Storer.SetReference(s.Ref)
	CheckIfError(err)
	for _, ref := range s.Summands {
		err = TheRepository.Storer.SetReference(ref)
		CheckIfError(err)
	}
}

type Base struct {
	Ref *plumbing.Reference
}

func (s Base) Children() []*plumbing.Reference {
	return []*plumbing.Reference{}
}

func (s Base) Name() string {
	return branchName(s.Ref.Name())
}


// method of ... but we cannot implement it here
// lift from regular ref into ... Segment/Sum/Base
func Convert(ref *plumbing.Reference) GitHierarchy {
	repository := TheRepository

	// todo: must be head, not tag or ...
	name := ref.Name().String()
	// assert(string.isPrefix(HeadPrefix, name)
	name, _ = strings.CutPrefix(name, HeadPrefix)

	if is, summands := isSum(name, repository); is {
		return Sum{ref, summands}

	} else if is, base := isSegment(name, repository); is {
		startHash, err1 := repository.Reference(plumbing.ReferenceName(segmentStart(name)), true)
		CheckIfError(err1)

		return Segment{ref, base, startHash}
	}

	// fmt.Println("it's a plain ref", ref)
	return Base{ref}
}

// ErrStop

// - not string ... what is required NewSet .. `Comparable'
// so duplicate Refs are not equal?
// repository.Reference should cache.

// so refname <--> GitHierarchy?
// man ^^^ on that?

// I want to accept

//  (node)
//  identity
//  neighbors

type adapter struct {
	gh GitHierarchy
}

func GetHierarchy(a graph.NodeExpander) GitHierarchy {
	return a.(adapter).gh
}
// invalid receiver type GitHierarchy (pointer or interface type)
// NodeIdentity
func (a adapter) NodeIdentity() string {
	// fmt.Println("NodeIdentity", s.n)
	return GetHierarchy(a).Name()
}


func (a adapter) NodePrepare() graph.NodeExpander {

	return adapter{Convert(a.gh.(Base).Ref )}
}


func (a adapter) NodeChildren() []graph.NodeExpander {
	refs := a.gh.Children()
	// []*plumbing.Reference

	var result = make([]graph.NodeExpander, len(refs))

	for i, x := range refs {
		result[i] = adapter{Base{x}}
	}

	return result
}


// dump the linear graph from a given top.
func WalkHierarchy(top *plumbing.Reference) (*[]graph.NodeExpander, *graph.Graph) {
	// Create a GitHierarchy object/ array
	g := adapter{Base{top}}

	// fmt.Println("walk: starting with node", g.NodeIdentity())

	vertices, incidenceGraph := graph.DiscoverGraph( &[]graph.NodeExpander{g})
	return vertices, incidenceGraph
}
