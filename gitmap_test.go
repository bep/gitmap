// Copyright 2024 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gitmap

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

type expectedGitInfo struct {
	abbreviatedHash string
	hash            string
	authorDate      string
	commitDate      string
}

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
	c := qt.New(t)

	gr, err := Map(Options{Repository: repository, Revision: revision})
	c.Assert(err, qt.IsNil)

	gm := gr.Files

	c.Assert(len(gm), qt.Equals, 11)

	assertFile(c, gm, "testfiles/d1/d1.txt",
		expectedGitInfo{
			"39120eb",
			"39120eb28a2f8a0312f9b45f91b6abb687b7fd3c",
			"2016-07-20",
			"2016-07-20",
		},
		expectedGitInfo{
			"4d9ad73",
			"4d9ad733fa40310607ebe9f700d59dcac93ace89",
			"2016-07-20",
			"2016-07-20",
		},
		expectedGitInfo{
			"9d1dc47",
			"9d1dc478eef267829831226d913a3ca249c489d4",
			"2016-07-19",
			"2016-07-19",
		},
	)

	assertFile(c, gm, "testfiles/d2/d2.txt",
		expectedGitInfo{
			"39120eb",
			"39120eb28a2f8a0312f9b45f91b6abb687b7fd3c",
			"2016-07-20",
			"2016-07-20",
		}, expectedGitInfo{
			"9d1dc47",
			"9d1dc478eef267829831226d913a3ca249c489d4",
			"2016-07-19",
			"2016-07-19",
		},
	)

	assertFile(c, gm, "testfiles/amended.txt",
		expectedGitInfo{
			"7d46b65",
			"7d46b653c9674510d808815c4c92c7dc10bedc16",
			"2019-05-23",
			"2019-05-25",
		},
	)

	assertFile(c, gm, "README.md",
		expectedGitInfo{
			"0b830e4",
			"0b830e458446fdb774b1688af9b402acf388d6ab",
			"2016-07-22",
			"2016-07-22",
		},
	)
}

func assertFile(
	t *qt.C,
	gm GitMap,
	filename string,
	expected ...expectedGitInfo,
) {
	var (
		gi *GitInfo
		ok bool
	)

	if gi, ok = gm[filename]; !ok {
		t.Fatal(filename)
	}

	for i, e := range expected {
		if i > 0 {
			if len(gi.Ancestors()) == 0 {
				t.Fatalf("Expected at least 1 ancestor commit for %s, but got none", filename)
			}
			gi = gi.Parent
			if gi == nil {
				t.Fatalf("Wrong number of ancestor commits, got %d, expected at least %d", i-1, i)
			}
		}
		assertGitInfo(t, *gi,
			filename,
			e.abbreviatedHash,
			e.hash,
			e.authorDate,
			e.commitDate,
		)
	}
}

func assertGitInfo(
	t *qt.C,
	gi GitInfo,
	filename string,
	expectedAbbreviatedHash,
	expectedHash,
	expectedAuthorDate,
	expectedCommitDate string,
) {
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

func TestCommitMessage(t *testing.T) {
	var (
		gm  GitMap
		gr  *GitRepo
		err error
	)

	if gr, err = Map(Options{Repository: repository, Revision: "HEAD"}); err != nil {
		t.Fatal(err)
	}

	gm = gr.Files

	assertMessage(
		t, gm,
		"testfiles/d1/d1.txt",
		"Change the test files",
		"To trigger a test variant.",
	)

	assertMessage(
		t, gm,
		"testfiles/r3.txt",
		"Edit testfiles/r3.txt",
		"Multiline\n\ncommit body.",
	)

	assertMessage(
		t, gm,
		"testfiles/amended.txt",
		"Add testfile with different author/commit date",
		"",
	)
}

func assertMessage(
	t *testing.T,
	gm GitMap,
	filename,
	expectedSubject,
	expectedBody string,
) {
	t.Helper()

	var (
		gi *GitInfo
		ok bool
	)

	if gi, ok = gm[filename]; !ok {
		t.Fatal(filename)
	}

	if gi.Subject != expectedSubject {
		t.Fatalf("Incorrect commit subject. Expected:\n%q\nGot:\n%q", expectedSubject, gi.Subject)
	}

	if gi.Body != expectedBody {
		t.Fatalf("Incorrect commit body. Expected:\n%q\nGot:\n%q", expectedBody, gi.Body)
	}
}

func TestActiveRevision(t *testing.T) {
	var (
		gm  GitMap
		gr  *GitRepo
		err error
	)

	if gr, err = Map(Options{Repository: repository, Revision: "HEAD"}); err != nil {
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
	gi, err := Map(Options{Repository: repository, Revision: revision})

	if err != ErrGitNotFound || gi != nil {
		t.Fatal("Invalid error handling")
	}
}

func TestEncodeJSON(t *testing.T) {
	const (
		filename = "README.md"
		revision = "1cb4bde80efbcc203ad14f8869c1fcca6ec830da"
	)

	c := qt.New(t)

	gi := getOne(c, revision, filename)

	b, err := json.Marshal(&gi)
	c.Assert(err, qt.IsNil)

	s := string(b)
	c.Assert(s, qt.Equals, `{"hash":"1cb4bde80efbcc203ad14f8869c1fcca6ec830da","abbreviatedHash":"1cb4bde","subject":"Add some badges to README","authorName":"Bjørn Erik Pedersen","authorEmail":"bjorn.erik.pedersen@gmail.com","authorDate":"2016-07-20T00:11:54+02:00","commitDate":"2016-07-20T00:11:54+02:00","body":""}`)
}

func TestAncestors(t *testing.T) {
	const (
		filename = "README.md"
		revision = "HEAD"
	)

	c := qt.New(t)

	gi := getOne(c, revision, filename)

	ancestors := gi.Ancestors()
	first := ancestors[0]
	last := ancestors[len(ancestors)-1]
	c.Assert(len(ancestors), qt.Equals, 5)
	c.Assert(first.Subject, qt.Equals, "Add some more to README")
	c.Assert(last.Subject, qt.Equals, "Create README.md")

	reversed := ancestors.Reverse()
	// Verify that the original slice is not modified.
	c.Assert(ancestors[0].Subject, qt.Equals, first.Subject)
	c.Assert(len(reversed), qt.Equals, 5)
	c.Assert(reversed[0].Subject, qt.Equals, last.Subject)
	c.Assert(reversed[len(reversed)-1].Subject, qt.Equals, first.Subject)
}

func TestGitRevisionNotFound(t *testing.T) {
	gi, err := Map(Options{Repository: repository, Revision: "adfasdfasdf"})

	// TODO(bep) improve error handling.
	if err == nil || gi != nil {
		t.Fatal("Invalid error handling", err)
	}
}

func TestGitRepoNotFound(t *testing.T) {
	gi, err := Map(Options{Repository: "adfasdfasdf", Revision: revision})

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

	if gr, err = Map(Options{Repository: repository, Revision: revision}); err != nil {
		t.Fatal(err)
	}

	expected := "/gitmap"

	if !strings.HasSuffix(gr.TopLevelAbsPath, expected) {
		t.Fatalf("Expected to end with %q got %q", expected, gr.TopLevelAbsPath)
	}
}

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Map(Options{Repository: repository, Revision: revision})
		if err != nil {
			b.Fatalf("Got error: %s", err)
		}
	}
}

func getOne(c *qt.C, revision, filename string) *GitInfo {
	c.Helper()
	gr, err := Map(Options{Repository: repository, Revision: revision})
	c.Assert(err, qt.IsNil)
	gi, ok := gr.Files[filename]
	c.Assert(ok, qt.IsTrue, qt.Commentf("File %s not found in GitMap", filename))
	return gi
}
