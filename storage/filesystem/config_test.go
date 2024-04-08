package filesystem

import (
	"github.com/go-git/go-billy/v5/osfs"
	fixtures "github.com/go-git/go-git-fixtures/v4"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/internal/test"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	. "gopkg.in/check.v1"
)

type ConfigSuite struct {
	test.Suite

	dir  *dotgit.DotGit
	path string
}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpTest(c *C) {
	tmp := c.MkDir()
	s.dir = dotgit.New(osfs.New(tmp))
	s.path = tmp
}

func (s *ConfigSuite) TestRemotes(c *C) {
	dir := dotgit.New(fixtures.Basic().ByTag(".git").One().DotGit())
	storer := &ConfigStorage{dir}

	cfg, err := storer.Config()
	c.Assert(err, IsNil)

	remotes := cfg.Remotes
	c.Assert(remotes, HasLen, 1)
	remote := remotes["origin"]
	c.Assert(remote.Name, Equals, "origin")
	c.Assert(remote.URLs, DeepEquals, []string{"https://github.com/git-fixtures/basic"})
	c.Assert(remote.Fetch, DeepEquals, []config.RefSpec{config.RefSpec("+refs/heads/*:refs/remotes/origin/*")})
}
