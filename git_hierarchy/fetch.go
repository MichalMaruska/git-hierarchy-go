package git_hierarchy

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// given refs/remotes/debian/unstable -> debian unstable
func SplitRemoteRef(refName plumbing.ReferenceName) (string, string) {

	// assert refName.IsRemote()
	prefix, _ := strings.CutSuffix(plumbing.RefRevParseRules[4], "%s")
	rest, _ := strings.CutPrefix(refName.String(), prefix)

	i := strings.Index(rest, "/")

	// TheRepository.Remotes()
	remote := rest[:i]
	remoteBranch := rest[i+1:]

	fmt.Println("remote:", remote)
	return remote, remoteBranch
}

// Given remote and branch name _on_ the remote (this is used ),
// get the local name for that. Maybe we should use refspec?
func remoteBranch(remote string, remoteRef plumbing.ReferenceName) *plumbing.Reference {
	name := plumbing.ReferenceName("refs/remotes/" + remote + "/" + branchName(remoteRef))

	// println("searching for", name)
	ref, err := TheRepository.Reference(name, true)
	CheckIfError(err, "resolving remote branch")
	return ref
}

func FetchUpstreamOf(ref *plumbing.Reference) {

	if refName := ref.Name(); refName.IsRemote() {
		fmt.Println("Fetching remote:", ref.Name())
		remote, remoteBranch := SplitRemoteRef(refName)

		// func (r *Remote) Fetch(o *FetchOptions) error
		gitRun("fetch", remote, remoteBranch)
		// gitRun("log", "--oneline",  old_head + ".." + "FETCH_HEAD")
		// git branch --force ${base#refs/heads/} $remote/$remote_branch

	} else if refName.IsBranch() {

		// it seems I must checkout for "git pull"
		fmt.Println("it's a branch, remote?")

		config, err := TheRepository.Storer.Config()
		// config.LoadConfig(config.LocalScope)
		//  GlobalScope  ... ~/.gitconfig ... is broken. [x.y] section
		// LocalScope should be read from the a ConfigStorer
		CheckIfError(err, "load config")
		config.Validate()

		branchInfo := config.Branches[branchName(ref.Name())] //  refName.String()
		if (branchInfo == nil) {
			println("it does not follow a remote branch")
			return
		}

		println("it follows a remote branch -- found info: ",
			branchInfo.Name, branchInfo.Remote, branchInfo.Merge)

		// is it on it?
		if RefsToSameCommit(ref, remoteBranch(branchInfo.Remote, branchInfo.Merge)) {
			println("so the remote branch is identical, let's fetch")
			// gitRun("remote","prune", br.Remote)
			gitRun("fetch", "--prune", "--progress", "--verbose",
				branchInfo.Remote, branchInfo.Merge.String() + ":" + refName.String())
			// gitRun("fetch", branchInfo.Remote, string(branchInfo.Merge) + ":")
		}
		// git fetch debian refs/heads/main:

		// plumbing.ReferenceName
		// look at config  merge
		// git config branch.base.merge
		// remote
		// func (r *Remote) List(o *ListOptions) (rfs []*plumbing.Reference, err error)

		// good idea!
		// w, err := r.Worktree()
		// func (r *Repository) Remotes() ([]*Remote, error)
		// w.Pull(&git.PullOptions{RemoteName: "origin"})
		// gitRun("pull")

	} else {
		fmt.Println("what is it?")
		// no way to modify
		// head
	}
	// rest
}


func fetch_upstream_ofs(nodes []GitHierarchy) {
	for _, gh := range nodes {
		switch gh.(type) {
		case  Base:
			FetchUpstreamOf(gh.(Base).Ref)
		default:
			// no
		}
	}
}
