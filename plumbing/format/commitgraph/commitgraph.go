package commitgraph

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

// CommitData is a reduced representation of Commit as presented in the commit graph
// file. It is merely useful as an optimization for walking the commit graphs.
//
// Deprecated: This package uses the wrong types for Generation and Index in CommitData.
// Use the v2 package instead.
type CommitData struct {
	// TreeHash is the hash of the root tree of the commit.
	TreeHash common.ObjectHash
	// ParentIndexes are the indexes of the parent commits of the commit.
	ParentIndexes []int
	// ParentHashes are the hashes of the parent commits of the commit.
	ParentHashes []common.ObjectHash
	// Generation number is the pre-computed generation in the commit graph
	// or zero if not available
	Generation int
	// When is the timestamp of the commit.
	When time.Time
}

// Index represents a representation of commit graph that allows indexed
// access to the nodes using commit object hash
//
// Deprecated: This package uses the wrong types for Generation and Index in CommitData.
// Use the v2 package instead.
type Index interface {
	// GetIndexByHash gets the index in the commit graph from commit hash, if available
	GetIndexByHash(h common.ObjectHash) (int, error)
	// GetNodeByIndex gets the commit node from the commit graph using index
	// obtained from child node, if available
	GetCommitDataByIndex(i int) (*CommitData, error)
	// Hashes returns all the hashes that are available in the index
	Hashes() []common.ObjectHash
}
