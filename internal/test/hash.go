package test

import "github.com/go-git/go-git/v5/plumbing/hash/common"

func X(h common.ObjectHash, _ bool) common.ObjectHash {
	return h
}
