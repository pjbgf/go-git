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
var algos = map[crypto.Hash]*common.SupportedHash{}

func init() {
	reset()
}

// reset resets the default algos value. Can be used after running tests
// that registers new algorithms to avoid side effects.
func reset() {
	algos[crypto.SHA1] = &common.SupportedHash{
		ObjectFormat: format.SHA1,
		NewHasher:    sha1cd.New,
		HashFactory:  sha1.SHA1HashFactory{},
	}
	algos[crypto.SHA256] = &common.SupportedHash{
		ObjectFormat: format.SHA256,
		NewHasher:    crypto.SHA256.New,
		HashFactory:  sha256.SHA256HashFactory{},
	}
}

// RegisterHash allows for the hash algorithm used to be overridden.
// This ensures the hash selection for go-git must be explicit, when
// overriding the default value.
func RegisterHash(h crypto.Hash, sh common.SupportedHash) error {
	if sh.NewHasher == nil {
		return fmt.Errorf("cannot register hash: NewHasher is nil")
	}

	algos[h] = &sh
	return nil
}

// Hash is the same as hash.Hash. This allows consumers
// to not having to import this package alongside "hash".
type Hash interface {
	hash.Hash
}

// New returns a new Hash for the given hash function.
// It panics if the hash function is not registered.
// Deprecated
func New(h crypto.Hash) Hash {
	hh, ok := algos[h]
	if !ok {
		panic(fmt.Sprintf("hash algorithm not registered: %v", h))
	}
	return hh.NewHasher()
}

// FromObjectFormat returns the correct Hash to be used based on the
// ObjectFormat being used.
// If the ObjectFormat is not recognised, returns ErrInvalidObjectFormat.
func FromObjectFormat(f format.ObjectFormat) (hash.Hash, error) {
	switch f {
	case format.SHA1:
		return New(crypto.SHA1), nil
	case format.SHA256:
		return New(crypto.SHA256), nil
	default:
		return nil, format.ErrInvalidObjectFormat
	}
}

// // FromHex parses a hexadecimal string and returns an Hash
// // and a boolean confirming whether the operation was successful.
// // The hash (and object format) is inferred from the length of the
// // input.
// //
// // Partial hashes will be handled as SHA1.
// func FromHex(in string) (Hash, bool) {
// 	switch len(in) {
// 	case hash.SHA256HexSize:
// 		return SHA256HashFromHex(in)
// 	default:
// 		return SHA1HashFromHex(in)
// 	}
// }

// // FromBytes creates an Hash object based on the value its
// // Sum() should return.
// // The hash type (and object format) is inferred from the length of
// // the input.
// //
// // Partial hashes will be handled as SHA1, and only the initial 20
// // bytes will be copied into the resulting hash.
// func FromBytes(in []byte) Hash {
// 	switch len(in) {
// 	case hash.SHA256Size:
// 		return SHA256HashFromBytes(in)
// 	default:
// 		return SHA1HashFromBytes(in)
// 	}
// }

// // ZeroFromHash returns a zeroed hash based on the given hash.Hash.
// //
// // Defaults to SHA1-sized hash if the provided hash is not supported.
// func ZeroFromHash(h hash.Hash) Hash {
// 	switch h.Size() {
// 	case hash.SHA256Size:
// 		return ZeroHashSHA256
// 	default:
// 		return ZeroHashSHA1
// 	}
// }

// // ZeroFromObjectFormat returns a zeroed hash based on the given ObjectFormat.
// //
// // Defaults to SHA1-sized hash if the provided format is not supported.
// func ZeroFromObjectFormat(f format.ObjectFormat) Hash {
// 	switch f {
// 	case format.SHA256:
// 		return ZeroHashSHA256
// 	default:
// 		return ZeroHashSHA1
// 	}
// }
