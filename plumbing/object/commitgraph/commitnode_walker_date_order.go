package commitgraph

import (
	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

// NewCommitNodeIterDateOrder returns a CommitNodeIter that walks the commit history,
// starting at the given commit and visiting its parents in Committer Time and Generation order,
// but with the  constraint that no parent is emitted before its children are emitted.
//
// This matches `git log --date-order`
func NewCommitNodeIterDateOrder(c CommitNode,
	seenExternal map[common.ObjectHash]bool,
	ignore []common.ObjectHash,
) CommitNodeIter {
	seen := make(map[common.ObjectHash]struct{})
	for _, h := range ignore {
		seen[h] = struct{}{}
	}
	for h, ext := range seenExternal {
		if ext {
			seen[h] = struct{}{}
		}
	}
	inCounts := make(map[common.ObjectHash]int)

	exploreHeap := &commitNodeHeap{binaryheap.NewWith(generationAndDateOrderComparator)}
	exploreHeap.Push(c)

	visitHeap := &commitNodeHeap{binaryheap.NewWith(generationAndDateOrderComparator)}
	visitHeap.Push(c)

	return &commitNodeIteratorTopological{
		exploreStack: exploreHeap,
		visitStack:   visitHeap,
		inCounts:     inCounts,
		ignore:       seen,
	}
}
