package sha1

import (
	"encoding/hex"

	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

var Factory = SHA1HashFactory{}

type SHA1HashFactory struct {
}

func (SHA1HashFactory) ZeroHash() common.ObjectHash {
	return zeroHash
}

func (SHA1HashFactory) FromHex(in string) (common.ObjectHash, bool) {
	return FromHex(in)
}

func (SHA1HashFactory) FromBytes(b []byte) common.ObjectHash {
	return FromBytes(b)
}

func (SHA1HashFactory) Size() int {
	return Size
}

func (SHA1HashFactory) HexSize() int {
	return HexSize
}

func ZeroHash() SHA1Hash {
	return zeroHash
}

func FromHex(in string) (SHA1Hash, bool) {
	if len(in) > HexSize {
		return zeroHash, false
	}

	b, err := hex.DecodeString(in)
	if err != nil {
		return zeroHash, false
	}

	h := SHA1Hash{}
	copy(h[:], b)
	return h, true
}

func FromBytes(b []byte) SHA1Hash {
	l := len(b)
	if l > Size {
		l = Size
	}

	h := SHA1Hash{}
	copy(h[:], b[:l])
	return h
}
