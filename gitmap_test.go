// Copyright © 2016-present Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gitmap_test

import (
	"os"
	"testing"

	"github.com/bep/gitmap"
)

func TestMap(t *testing.T) {
	var (
		repository string
		gm         gitmap.GitMap
		err        error
	)

	if repository, err = os.Getwd(); err != nil {
		t.Fatal(err)
	}

	if gm, err = gitmap.Map(repository, "37e91d4"); err != nil {
		t.Fatal(err)
	}

	if len(gm) != 8 {
		t.Fatalf("Wrong number of files, got %d, expected %d", len(gm), 8)
	}

	assertFile(t, gm,
		"testfiles/d1/d1.txt",
		"37e91d4",
		"37e91d49494bd894e4086565d2c3ab8c6351820e",
	)

	assertFile(t, gm,
		"README.md",
		"527cb5d",
		"527cb5db32c76a269e444bb0de4cc22b574f0366",
	)
}

func assertFile(
	t *testing.T,
	gm gitmap.GitMap,
	filename,
	expectedAbbreviatedTreeHash,
	expectedTreeHash string) {

	var (
		gi *gitmap.GitInfo
		ok bool
	)

	if gi, ok = gm[filename]; !ok {
		t.Fatalf(filename)
	}

	if gi.AbbreviatedHash != expectedAbbreviatedTreeHash || gi.Hash != expectedTreeHash {
		t.Error("Invalid tree hash, file", filename, "abbreviated:", gi.AbbreviatedHash, "full:", gi.Hash, gi.Subject)
	}

	if gi.AuthorName != "Bjørn Erik Pedersen" || gi.AuthorEmail != "bjorn.erik.pedersen@gmail.com" {
		t.Error("These commits are mine! Got", gi.AuthorName, "and", gi.AuthorEmail)
	}
}
