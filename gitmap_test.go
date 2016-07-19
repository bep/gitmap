// Copyright © 2016-present Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gitmap

import (
	"os"
	"testing"
)

var (
	revision   = "9d1dc47"
	repository string
)

func init() {
	var err error
	if repository, err = os.Getwd(); err != nil {
		panic(err)
	}
}

func TestMap(t *testing.T) {
	var (
		gm  GitMap
		err error
	)

	if gm, err = Map(repository, revision); err != nil {
		t.Fatal(err)
	}

	if len(gm) != 8 {
		t.Fatalf("Wrong number of files, got %d, expected %d", len(gm), 8)
	}

	assertFile(t, gm,
		"testfiles/d1/d1.txt",
		"9d1dc47",
		"9d1dc478eef267829831226d913a3ca249c489d4",
	)

	assertFile(t, gm,
		"README.md",
		"527cb5d",
		"527cb5db32c76a269e444bb0de4cc22b574f0366",
	)
}

func assertFile(
	t *testing.T,
	gm GitMap,
	filename,
	expectedAbbreviatedHash,
	expectedHash string) {

	var (
		gi *GitInfo
		ok bool
	)

	if gi, ok = gm[filename]; !ok {
		t.Fatalf(filename)
	}

	if gi.AbbreviatedHash != expectedAbbreviatedHash || gi.Hash != expectedHash {
		t.Error("Invalid tree hash, file", filename, "abbreviated:", gi.AbbreviatedHash, "full:", gi.Hash, gi.Subject)
	}

	if gi.AuthorName != "Bjørn Erik Pedersen" || gi.AuthorEmail != "bjorn.erik.pedersen@gmail.com" {
		t.Error("These commits are mine! Got", gi.AuthorName, "and", gi.AuthorEmail)
	}

	if gi.AuthorDate.Format("2006-01-02") != "2016-07-19" {
		t.Error("Invalid date:", gi.AuthorDate)
	}
}

func TestGitExecutableNotFound(t *testing.T) {
	defer initDefaults()
	gitExec = "thisShouldHopefullyNotExistOnPath"
	gi, err := Map(repository, revision)

	if err != GitNotFound || gi != nil {
		t.Fatal("Invalid error handling")
	}

}

func TestGitRevisionNotFound(t *testing.T) {
	gi, err := Map(repository, "adfasdfasdf")

	// TODO(bep) improve error handling.
	if err == nil || gi != nil {
		t.Fatal("Invalid error handling", err)
	}
}

func TestGitRepoNotFound(t *testing.T) {
	gi, err := Map("adfasdfasdf", revision)

	// TODO(bep) improve error handling.
	if err == nil || gi != nil {
		t.Fatal("Invalid error handling", err)
	}
}
