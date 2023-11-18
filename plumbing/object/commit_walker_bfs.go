package object

import (
	"io"

	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type bfsCommitIterator struct {
	seenExternal map[common.ObjectHash]bool
	seen         map[common.ObjectHash]bool
	queue        []*Commit
}

// NewCommitIterBSF returns a CommitIter that walks the commit history,
// starting at the given commit and visiting its parents in pre-order.
// The given callback will be called for each visited commit. Each commit will
// be visited only once. If the callback returns an error, walking will stop
// and will return the error. Other errors might be returned if the history
// cannot be traversed (e.g. missing objects). Ignore allows to skip some
// commits from being iterated.
func NewCommitIterBSF(
	c *Commit,
	seenExternal map[common.ObjectHash]bool,
	ignore []common.ObjectHash,
) CommitIter {
	seen := make(map[common.ObjectHash]bool)
	for _, h := range ignore {
		seen[h] = true
	}

	return &bfsCommitIterator{
		seenExternal: seenExternal,
		seen:         seen,
		queue:        []*Commit{c},
	}
}

func (w *bfsCommitIterator) appendHash(store storer.EncodedObjectStorer, h common.ObjectHash) error {
	if w.seen[h] || w.seenExternal[h] {
		return nil
	}
	c, err := GetCommit(store, h)
	if err != nil {
		return err
	}
	w.queue = append(w.queue, c)
	return nil
}

func (w *bfsCommitIterator) Next() (*Commit, error) {
	var c *Commit
	for {
		if len(w.queue) == 0 {
			return nil, io.EOF
		}
		c = w.queue[0]
		w.queue = w.queue[1:]

		if w.seen[c.Hash] || w.seenExternal[c.Hash] {
			continue
		}

		w.seen[c.Hash] = true

		for _, h := range c.ParentHashes {
			err := w.appendHash(c.s, h)
			if err != nil {
				return nil, err
			}
		}

		return c, nil
	}
}

func (w *bfsCommitIterator) ForEach(cb func(*Commit) error) error {
	for {
		c, err := w.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *bfsCommitIterator) Close() {}
