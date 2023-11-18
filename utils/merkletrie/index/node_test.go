package index

import (
	"bytes"
	"path/filepath"
	"testing"

	. "github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type NoderSuite struct{}

var _ = Suite(&NoderSuite{})

func (s *NoderSuite) TestDiff(c *C) {
	indexA := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/qux", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/baz/foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
		},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/qux", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "bar/baz/foo", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
		},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	c.Assert(err, IsNil)
	c.Assert(ch, HasLen, 0)
}

func (s *NoderSuite) TestDiffChange(c *C) {
	indexA := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("bar", "baz", "bar"),
			Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d")),
		}},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("bar", "baz", "foo"),
			Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d")),
		}},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	c.Assert(err, IsNil)
	c.Assert(ch, HasLen, 2)
}

func (s *NoderSuite) TestDiffDir(c *C) {
	indexA := &index.Index{
		Entries: []*index.Entry{{
			Name: "foo",
			Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d")),
		}},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{{
			Name: filepath.Join("foo", "bar"),
			Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d")),
		}},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	c.Assert(err, IsNil)
	c.Assert(ch, HasLen, 2)
}

func (s *NoderSuite) TestDiffSameRoot(c *C) {
	indexA := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo.go", Hash: X(sha1.FromHex("aab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "foo/bar", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
		},
	}

	indexB := &index.Index{
		Entries: []*index.Entry{
			{Name: "foo/bar", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
			{Name: "foo.go", Hash: X(sha1.FromHex("8ab686eafeb1f44702738c8b0f24f2567c36da6d"))},
		},
	}

	ch, err := merkletrie.DiffTree(NewRootNode(indexA), NewRootNode(indexB), isEquals)
	c.Assert(err, IsNil)
	c.Assert(ch, HasLen, 1)
}

var empty = make([]byte, 24)

func isEquals(a, b noder.Hasher) bool {
	if bytes.Equal(a.Hash(), empty) || bytes.Equal(b.Hash(), empty) {
		return false
	}

	return bytes.Equal(a.Hash(), b.Hash())
}
