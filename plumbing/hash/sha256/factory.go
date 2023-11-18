package sha256

import (
	"encoding/hex"

	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

var Factory = SHA256HashFactory{}

type SHA256HashFactory struct {
}

func (SHA256HashFactory) ZeroHash() common.ObjectHash {
	return zeroHash
}

func (SHA256HashFactory) FromHex(in string) (common.ObjectHash, bool) {
	return FromHex(in)
}

func (SHA256HashFactory) FromBytes(b []byte) common.ObjectHash {
	return FromBytes(b)
}

func (SHA256HashFactory) Size() int {
	return Size
}

func (SHA256HashFactory) HexSize() int {
	return HexSize
}

func ZeroHash() SHA256Hash {
	return zeroHash
}

func FromHex(in string) (SHA256Hash, bool) {
	if len(in) > HexSize {
		return zeroHash, false
	}

	b, err := hex.DecodeString(in)
	if err != nil {
		return zeroHash, false
	}

	h := SHA256Hash{}
	copy(h[:], b)
	return h, true
}

func FromBytes(b []byte) SHA256Hash {
	l := len(b)
	if len(b) > Size {
		l = Size
	}

	h := SHA256Hash{}
	copy(h[:], b[:l])
	return h
}
