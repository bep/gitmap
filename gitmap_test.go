// Copyright © 2016-present Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gitmap

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

var (
	revision   = "7d46b653c9674510d808815c4c92c7dc10bedc16"
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
		gr  *GitRepo
		err error
	)

	if gr, err = Map(repository, revision); err != nil {
		t.Fatal(err)
	}

	gm = gr.Files

	if len(gm) != 11 {
		t.Fatalf("Wrong number of files, got %d, expected %d", len(gm), 9)
	}

	assertFile(t, gm,
		"testfiles/d1/d1.txt",
		"39120eb",
		"39120eb28a2f8a0312f9b45f91b6abb687b7fd3c",
		"2016-07-20",
		"2016-07-20",
	)

	assertFile(t, gm,
		"testfiles/d2/d2.txt",
		"39120eb",
		"39120eb28a2f8a0312f9b45f91b6abb687b7fd3c",
		"2016-07-20",
		"2016-07-20",
	)

	assertFile(t, gm,
		"testfiles/amended.txt",
		"7d46b65",
		"7d46b653c9674510d808815c4c92c7dc10bedc16",
		"2019-05-23",
		"2019-05-25",
	)

	assertFile(t, gm,
		"README.md",
		"0b830e4",
		"0b830e458446fdb774b1688af9b402acf388d6ab",
		"2016-07-22",
		"2016-07-22",
	)
}

func assertFile(
	t *testing.T,
	gm GitMap,
	filename,
	expectedAbbreviatedHash,
	expectedHash,
	expectedAuthorDate,
	expectedCommitDate string) {

	var (
		gi *GitInfo
		ok bool
	)

	if gi, ok = gm[filename]; !ok {
		t.Fatal(filename)
	}

	if gi.AbbreviatedHash != expectedAbbreviatedHash || gi.Hash != expectedHash {
		t.Error("Invalid tree hash, file", filename, "abbreviated:", gi.AbbreviatedHash, "full:", gi.Hash, gi.Subject)
	}

	if gi.AuthorName != "Bjørn Erik Pedersen" && gi.AuthorName != "Michael Stapelberg" {
		t.Error("These commits are mine! Got", gi.AuthorName, "and", gi.AuthorEmail)
	}

	if gi.AuthorEmail != "bjorn.erik.pedersen@gmail.com" && gi.AuthorEmail != "stapelberg@google.com" {
		t.Error("These commits are mine! Got", gi.AuthorName, "and", gi.AuthorEmail)
	}

	if got, want := gi.AuthorDate.Format("2006-01-02"), expectedAuthorDate; got != want {
		t.Errorf("%s: unexpected author date: got %v, want %v", filename, got, want)
	}

	if got, want := gi.CommitDate.Format("2006-01-02"), expectedCommitDate; got != want {
		t.Errorf("%s: unexpected commit date: got %v, want %v", filename, got, want)
	}
}

func TestActiveRevision(t *testing.T) {
	var (
		gm  GitMap
		gr  *GitRepo
		err error
	)

	if gr, err = Map(repository, "HEAD"); err != nil {
		t.Fatal(err)
	}

	gm = gr.Files

	if len(gm) < 10 {
		t.Fatalf("Wrong number of files, got %d, expected at least %d", len(gm), 10)
	}

	if len(gm) < 10 {
		t.Fatalf("Wrong number of files, got %d, expected at least %d", len(gm), 10)
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

func TestEncodeJSON(t *testing.T) {
	var (
		gm       GitMap
		gr       *GitRepo
		gi       *GitInfo
		err      error
		ok       bool
		filename = "README.md"
	)

	if gr, err = Map(repository, revision); err != nil {
		t.Fatal(err)
	}

	gm = gr.Files

	if gi, ok = gm[filename]; !ok {
		t.Fatal(filename)
	}

	b, err := json.Marshal(&gi)

	if err != nil {
		t.Fatal(err)
	}

	s := string(b)

	if s != `{"hash":"0b830e458446fdb774b1688af9b402acf388d6ab","abbreviatedHash":"0b830e4","subject":"Add some more to README","authorName":"Bjørn Erik Pedersen","authorEmail":"bjorn.erik.pedersen@gmail.com","authorDate":"2016-07-22T21:40:27+02:00","commitDate":"2016-07-22T21:40:27+02:00"}` {
		t.Errorf("JSON marshal error: \n%s", s)
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

func TestTopLevelAbsPath(t *testing.T) {
	var (
		gr  *GitRepo
		err error
	)

	if gr, err = Map(repository, revision); err != nil {
		t.Fatal(err)
	}

	expected := "/bep/gitmap"

	if !strings.HasSuffix(gr.TopLevelAbsPath, expected) {
		t.Fatalf("Expected to end with %q got %q", expected, gr.TopLevelAbsPath)
	}
}

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Map(repository, revision)
		if err != nil {
			b.Fatalf("Got error: %s", err)
		}
	}
}
