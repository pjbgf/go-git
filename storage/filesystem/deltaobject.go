package filesystem

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

type deltaObject struct {
	plumbing.EncodedObject
	base common.ObjectHash
	hash common.ObjectHash
	size int64
}

func newDeltaObject(
	obj plumbing.EncodedObject,
	hash common.ObjectHash,
	base common.ObjectHash,
	size int64) plumbing.DeltaObject {
	return &deltaObject{
		EncodedObject: obj,
		hash:          hash,
		base:          base,
		size:          size,
	}
}

func (o *deltaObject) BaseHash() common.ObjectHash {
	return o.base
}

func (o *deltaObject) ActualSize() int64 {
	return o.size
}

func (o *deltaObject) ActualHash() common.ObjectHash {
	return o.hash
}
