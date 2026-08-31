package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BranchPrefix namespaces every branch gh-wip creates.
const BranchPrefix = "wip/"

// NewBranchName generates a sorting-friendly WIP branch name from t.
func NewBranchName(t time.Time) string {
	return BranchPrefix + t.UTC().Format("20060102-150405")
}

// WipBranch describes a remote WIP branch and the metadata shown in the
// interactive picker during `gh wip pull`.
type WipBranch struct {
	// Name is the short branch name, e.g. "wip/20260831-140501".
	Name string
	// Subject is the summary line of its tip commit.
	Subject string
	// Author is the tip commit's author name.
	Author string
	// When is the tip commit's author date.
	When time.Time
}

// Ref returns the branch's fully-qualified remote-tracking ref, e.g.
// "origin/wip/20260831-140501".
func (b WipBranch) Ref(remote string) string {
	return remote + "/" + b.Name
}

// ListWipBranches returns every remote-tracking WIP branch for remote,
// newest first. Callers should Fetch(remote) beforehand to see the latest.
func ListWipBranches(remote string) ([]WipBranch, error) {
	// %00-delimited fields, %09 between refname/subject/author/date so a
	// commit subject containing a literal field separator can't corrupt
	// parsing.
	const format = "%(refname:short)%09%(subject)%09%(authorname)%09%(authordate:unix)"
	out, err := run("for-each-ref", "--sort=-committerdate", "--format="+format, "refs/remotes/"+remote+"/"+BranchPrefix+"*")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	prefix := remote + "/"
	var branches []WipBranch
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			continue
		}
		short := strings.TrimPrefix(fields[0], prefix)
		unix, _ := strconv.ParseInt(fields[3], 10, 64)
		branches = append(branches, WipBranch{
			Name:    short,
			Subject: fields[1],
			Author:  fields[2],
			When:    time.Unix(unix, 0),
		})
	}
	return branches, nil
}

// UniqueBranchName returns name, or name with a numeric suffix appended
// until it no longer collides with an existing local branch. Collisions are
// only realistically possible when gh-wip is invoked more than once within
// the same second.
func UniqueBranchName(name string) string {
	if !LocalBranchExists(name) {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", name, i)
		if !LocalBranchExists(candidate) {
			return candidate
		}
	}
}
