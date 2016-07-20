// Copyright © 2016-present Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gitmap

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	// will be modified during tests
	gitExec string

	GitNotFound = errors.New("Git executable not found in $PATH")
)

// GitMap maps filenames to Git revision information.
type GitMap map[string]*GitInfo

// GitInfo holds information about a Git commit.
type GitInfo struct {
	Hash            string    `json:"hash"`            // Commit hash
	AbbreviatedHash string    `json:"abbreviatedHash"` // Abbreviated commit hash
	Subject         string    `json:"subject"`         // The commit message's subject/title line
	AuthorName      string    `json:"authorName"`      // The author name, respecting .mailmap
	AuthorEmail     string    `json:"authorEmail"`     // The author email address, respecting .mailmap
	AuthorDate      time.Time `json:"authorDate"`      // The author date
}

// Map creates a GitMap from the given repository path and revision.
// Use blank or HEAD as revision for the currently active revision.
func Map(repository, revision string) (GitMap, error) {
	m := make(GitMap)

	gitLogArgs := fmt.Sprintf(
		"-C %s log --name-only --reverse --format=format:%%x1e%%H%%x1f%%h%%x1f%%s%%x1f%%aN%%x1f%%aE%%x1f%%ai %s",
		repository,
		revision,
	)

	out, err := exec.Command(gitExec, strings.Fields(gitLogArgs)...).Output()

	if err != nil {
		if ee, ok := err.(*exec.Error); ok {
			if ee.Err == exec.ErrNotFound {
				return nil, GitNotFound
			}
		}

		return nil, err
	}

	entriesStr := string(out)
	entriesStr = strings.Trim(entriesStr, "\n\x1e'")
	entries := strings.Split(entriesStr, "\x1e")

	for _, e := range entries {
		lines := strings.Split(e, "\n")
		gitInfo, err := toGitInfo(lines[0])

		if err != nil {
			return nil, err
		}

		for _, filename := range lines[1:] {
			filename := strings.TrimSpace(filename)
			if filename == "" {
				continue
			}
			m[filename] = gitInfo
		}
	}

	return m, nil
}

func toGitInfo(entry string) (*GitInfo, error) {
	items := strings.Split(entry, "\x1f")
	authorDate, err := time.Parse("2006-01-02 15:04:05 -0700", items[5])

	if err != nil {
		return nil, err
	}

	return &GitInfo{
		Hash:            items[0],
		AbbreviatedHash: items[1],
		Subject:         items[2],
		AuthorName:      items[3],
		AuthorEmail:     items[4],
		AuthorDate:      authorDate,
	}, nil
}

func init() {
	initDefaults()
}

func initDefaults() {
	gitExec = "git"
}
