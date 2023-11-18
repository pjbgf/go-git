package transactional

import (
	. "github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/storage/memory"

	. "gopkg.in/check.v1"
)

var _ = Suite(&ShallowSuite{})

type ShallowSuite struct{}

func (s *ShallowSuite) TestShallow(c *C) {
	base := memory.NewStorage()
	temporal := memory.NewStorage()

	rs := NewShallowStorage(base, temporal)

	commitA := X(sha1.FromHex("bc9968d75e48de59f0870ffb71f5e160bbbdcf52"))
	commitB := X(sha1.FromHex("aa9968d75e48de59f0870ffb71f5e160bbbdcf52"))

	err := base.SetShallow([]common.ObjectHash{commitA})
	c.Assert(err, IsNil)

	err = rs.SetShallow([]common.ObjectHash{commitB})
	c.Assert(err, IsNil)

	commits, err := rs.Shallow()
	c.Assert(err, IsNil)
	c.Assert(commits, HasLen, 1)
	c.Assert(commits[0], Equals, commitB)

	commits, err = base.Shallow()
	c.Assert(err, IsNil)
	c.Assert(commits, HasLen, 1)
	c.Assert(commits[0], Equals, commitA)
}

func (s *ShallowSuite) TestCommit(c *C) {
	base := memory.NewStorage()
	temporal := memory.NewStorage()

	rs := NewShallowStorage(base, temporal)

	commitA := X(sha1.FromHex("bc9968d75e48de59f0870ffb71f5e160bbbdcf52"))
	commitB := X(sha1.FromHex("aa9968d75e48de59f0870ffb71f5e160bbbdcf52"))

	c.Assert(base.SetShallow([]common.ObjectHash{commitA}), IsNil)
	c.Assert(rs.SetShallow([]common.ObjectHash{commitB}), IsNil)

	c.Assert(rs.Commit(), IsNil)

	commits, err := rs.Shallow()
	c.Assert(err, IsNil)
	c.Assert(commits, HasLen, 1)
	c.Assert(commits[0], Equals, commitB)

	commits, err = base.Shallow()
	c.Assert(err, IsNil)
	c.Assert(commits, HasLen, 1)
	c.Assert(commits[0], Equals, commitB)
}
