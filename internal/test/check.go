package test

import (
	fixtures "github.com/go-git/go-git-fixtures/v4"
	"gopkg.in/check.v1"
)

// Suite replaces fixtures.Suite, now that check.v1 is deprecated on
// other go-git components.
//
// Deprecated: check.v1 should be replaced with stretchr/testify/assert.
type Suite struct{}

// Run executes test for each fixture in f.
func (s *Suite) Run(c *check.C, f fixtures.Fixtures, test func(*fixtures.Fixture)) {
	if c == nil {
		panic("c cannot be nil")
	}

	if f == nil {
		c.Fatal("fixtures cannot be nil")
	}

	if test == nil {
		c.Fatal("test cannot be nil")
	}

	for _, fix := range f {
		c.Logf("executing test at %s %s", fix.URL, fix.Tags)
		test(fix)
	}
}
