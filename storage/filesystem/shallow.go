package filesystem

import (
	"bufio"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/utils/ioutil"
)

// ShallowStorage where the shallow commits are stored, an internal to
// manipulate the shallow file
type ShallowStorage struct {
	dir *dotgit.DotGit
}

// SetShallow save the shallows in the shallow file in the .git folder as one
// commit per line represented by 40-byte hexadecimal object terminated by a
// newline.
func (s *ShallowStorage) SetShallow(commits []common.ObjectHash) error {
	f, err := s.dir.ShallowWriter()
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(f, &err)
	for _, h := range commits {
		if _, err := fmt.Fprintf(f, "%s\n", h); err != nil {
			return err
		}
	}

	return err
}

// Shallow returns the shallow commits reading from shallo file from .git
func (s *ShallowStorage) Shallow() ([]common.ObjectHash, error) {
	f, err := s.dir.Shallow()
	if f == nil || err != nil {
		return nil, err
	}

	defer ioutil.CheckClose(f, &err)

	var hash []common.ObjectHash

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		h, _ := sha1.FromHex(scn.Text())
		hash = append(hash, h)
	}

	return hash, scn.Err()
}
