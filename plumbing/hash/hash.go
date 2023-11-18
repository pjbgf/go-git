// package hash provides a way for managing the
// underlying hash implementations used across go-git.
package hash

import (
	"crypto"
	"errors"
	"fmt"
	"hash"

	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/hash/sha256"
	"github.com/pjbgf/sha1cd"
)

var (
	ErrUnsupportedHashFunction = errors.New("unsupported hash function")
)

// algos is a map of hash algorithms.
var algos = map[format.ObjectFormat]*common.SupportedHash{}

func init() {
	reset()
}

// reset resets the default algos value. Can be used after running tests
// that registers new algorithms to avoid side effects.
func reset() {
	algos[format.SHA1] = &common.SupportedHash{
		NewHasher:     sha1cd.New,
		ObjectFactory: sha1.Factory,
	}
	algos[format.SHA256] = &common.SupportedHash{
		NewHasher:     crypto.SHA256.New,
		ObjectFactory: sha256.Factory,
	}
}

// RegisterHash allows for the hash algorithm used to be overridden.
// This ensures the hash selection for go-git must be explicit, when
// overriding the default value.
func RegisterHash(f format.ObjectFormat, sh common.SupportedHash) error {
	if sh.NewHasher == nil {
		return fmt.Errorf("cannot register hash: NewHasher is nil")
	}

	algos[f] = &sh
	return nil
}

// Hash is the same as hash.Hash. This allows consumers
// to not having to import this package alongside "hash".
type Hash interface {
	hash.Hash
}

// New returns a new Hash for the given hash function.
// It panics if the hash function is not registered.
func NewHasher(f format.ObjectFormat) Hash {
	hh, ok := algos[f]
	if !ok {
		panic(fmt.Sprintf("hash algorithm not registered: %v", f))
	}
	return hh.NewHasher()
}

// New returns a new Hash for the given hash function.
// It panics if the hash function is not registered.
func HashFactory(f format.ObjectFormat) common.HashFactory {
	s, ok := algos[f]
	if !ok {
		panic(fmt.Sprintf("hash algorithm not registered: %v", f))
	}
	return s.ObjectFactory
}
