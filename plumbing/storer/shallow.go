package storer

import "github.com/go-git/go-git/v5/plumbing/hash/common"

// ShallowStorer is a storage of references to shallow commits by hash,
// meaning that these commits have missing parents because of a shallow fetch.
type ShallowStorer interface {
	SetShallow([]common.ObjectHash) error
	Shallow() ([]common.ObjectHash, error)
}
