package repo

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"slices"

	"github.com/gitsight/go-vcsurl"
)

// IsInRepo determines if the provided directory is in the provided git
// repository. Git repositories can be named differently on a local clone then
// the remote. In addition, a git repo can have multiple remotes. In practice
// though, the business logic question is something like:
// "Am I in the community repo?". To answer that, this function iterates over
// the remotes and if any of them have the same name as the one requested, it
// returns true. Any other case returns false.
func IsInRepo(dir string, names ...string) bool {
	if dir == "" {
		dir = "."
	}

	cmd := exec.Command("git", "-C", dir, "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	for line := range bytes.Lines(out) {
		fields := bytes.Fields(bytes.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}

		u, err := vcsurl.Parse(string(fields[1]))
		if err != nil {
			continue
		}

		if slices.Contains(names, u.Name) {
			return true
		}
	}

	return false
}

func Root(dir string) (string, error) {
	if dir == "" {
		dir = "."
	}

	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if err, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr = err.Stderr
		}
		return "", fmt.Errorf("failed to get repo root: %w: %s", err, string(stderr))
	}

	return string(bytes.TrimSpace(out)), nil
}
