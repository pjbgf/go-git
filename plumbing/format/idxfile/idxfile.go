package idxfile

import (
	"io"
	"sort"
	"sync"

	encbin "encoding/binary"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
)

// Version represents the version of the pack index file.
type Version uint32

const (
	// VersionNotSet
	// TODO: check upstream potential issues with this approach
	VersionNotSet Version = 0
	// Version2
	Version2 Version = 2
	// Version3
	Version3 Version = 3

	noMapping = -1

	isO64Mask = uint64(1) << 31
)

var (
	//  Note: this header isn't active yet.  In future versions of git
	//  we may change the index file format.  At that time we would start
	//  the first four bytes of the new index format with this signature,
	//  as all older git binaries would find this value illegal and abort
	//  reading the file.
	//
	//  This is the case because the number of objects in a packfile
	//  cannot exceed 1,431,660,000 as every object would need at least
	//  3 bytes of data and the overall packfile cannot exceed 4 GiB due
	//  to the 32 bit offsets used by the index.  Clearly the signature
	//  exceeds this maximum.
	//
	//  Very old git binaries will also compare the first 4 bytes to the
	//  next 4 bytes in the index and abort with a "non-monotonic index"
	//  error if the second 4 byte word is smaller than the first 4
	//  byte word.  This would be true in the proposed future index
	//  format as idx_signature would be greater than idx_version.
	//
	idxHeader = []byte{255, 't', 'O', 'c'} // \377tOc -> 0xff744f63
)

// Index represents an index of a packfile.
type Index interface {
	// Contains checks whether the given hash is in the index.
	Contains(h common.ObjectHash) (bool, error)
	// FindOffset finds the offset in the packfile for the object with
	// the given hash.
	FindOffset(h common.ObjectHash) (int64, error)
	// FindCRC32 finds the CRC32 of the object with the given hash.
	FindCRC32(h common.ObjectHash) (uint32, error)
	// FindHash finds the hash for the object with the given offset.
	FindHash(o int64) (common.ObjectHash, error)
	// Count returns the number of entries in the index.
	Count() (int64, error)
	// Entries returns an iterator to retrieve all index entries.
	Entries() (EntryIter, error)
	// EntriesByOffset returns an iterator to retrieve all index entries ordered
	// by offset.
	EntriesByOffset() (EntryIter, error)
}

// MemoryIndex is the in memory representation of an idx file.
type MemoryIndex struct {
	Version Version
	Fanout  [256]uint32
	// FanoutMapping maps the position in the fanout table to the position
	// in the Names, Offset32 and CRC32 slices. This improves the memory
	// usage by not needing an array with unnecessary empty slots.
	FanoutMapping [256]int
	// Names is a table of sorted object names. These are packed together
	// without offset values to reduce the cache footprint of the binary
	// search for a specific object name.
	Names [][]byte
	// Offset32 is a table of 4-byte offset values (in network byte order).
	// These are usually 31-bit pack file offsets, but large offsets are
	// encoded as an index into the next table with the msbit set.
	Offset32 [][]byte
	// CRC32 is a table of 4-byte CRC32 values of the packed object data.
	// This is new in v2 so compressed data can be copied directly from
	// pack to pack during repacking without undetected data corruption.
	CRC32 [][]byte
	// Offset64 is a table of 8-byte offset entries (empty for pack files
	// less than 2 GiB). Pack files are organized with heavily used objects
	// toward the front, so most object references should not need to refer
	// to this table.
	Offset64         []byte
	PackfileChecksum []byte
	IdxChecksum      []byte

	offsetHash       map[int64]common.ObjectHash
	offsetHashIsFull bool

	factory common.HashFactory
	m       sync.RWMutex
	// allocate header space once.
	header [4]byte
}

// NewMemoryIndex returns an instance of a new MemoryIndex.
func NewMemoryIndex(f common.HashFactory) *MemoryIndex {
	idx := &MemoryIndex{
		factory: f,
	}

	idx.Fanout = [256]uint32{}
	idx.FanoutMapping = [256]int{}
	idx.Names = make([][]byte, 0)
	idx.Offset32 = make([][]byte, 0)
	idx.CRC32 = make([][]byte, 0)
	idx.Offset64 = nil
	idx.PackfileChecksum = make([]byte, f.Size())
	idx.IdxChecksum = make([]byte, f.Size())
	idx.offsetHash = make(map[int64]common.ObjectHash)

	return idx
}

func (idx *MemoryIndex) reset(factory common.HashFactory) {
	// Clear out fanout.
	for i := range idx.Fanout {
		idx.Fanout[i] = 0
	}
	for i := range idx.header {
		idx.header[i] = 0
	}

	// Unmap all fans as a starting point.
	for i := range idx.FanoutMapping {
		idx.FanoutMapping[i] = noMapping
	}

	for i := range idx.Names {
		idx.Names[i] = idx.Names[i][:0]
	}
	idx.Names = idx.Names[:0]

	for i := range idx.Offset32 {
		idx.Offset32[i] = idx.Offset32[i][:0]
	}
	idx.Offset32 = idx.Offset32[:0]

	for i := range idx.CRC32 {
		idx.CRC32[i] = idx.CRC32[i][:0]
	}
	idx.CRC32 = idx.CRC32[:0]

	idx.Offset64 = idx.Offset64[:0]
	idx.PackfileChecksum = idx.PackfileChecksum[:0]
	idx.IdxChecksum = idx.IdxChecksum[:0]

	idx.offsetHash = make(map[int64]common.ObjectHash)
	idx.offsetHashIsFull = false
	idx.factory = factory
}

func (idx *MemoryIndex) findHashIndex(h common.ObjectHash) (int, bool) {
	k := idx.FanoutMapping[h.Sum()[0]]
	if k == noMapping {
		return 0, false
	}

	if len(idx.Names) <= k {
		return 0, false
	}

	data := idx.Names[k]
	high := uint64(len(idx.Offset32[k])) >> 2
	if high == 0 {
		return 0, false
	}

	low := uint64(0)
	for {
		mid := (low + high) >> 1
		objectIDLength := uint64(idx.factory.Size())
		offset := mid * objectIDLength

		cmp := h.Compare(data[offset : offset+objectIDLength])
		if cmp < 0 {
			high = mid
		} else if cmp == 0 {
			return int(mid), true
		} else {
			low = mid + 1
		}

		if low >= high {
			break
		}
	}

	return 0, false
}

// Contains implements the Index interface.
func (idx *MemoryIndex) Contains(h common.ObjectHash) (bool, error) {
	_, ok := idx.findHashIndex(h)
	return ok, nil
}

// FindOffset implements the Index interface.
func (idx *MemoryIndex) FindOffset(h common.ObjectHash) (int64, error) {
	if len(idx.FanoutMapping) <= int(h.Sum()[0]) {
		return 0, plumbing.ErrObjectNotFound
	}

	k := idx.FanoutMapping[h.Sum()[0]]
	i, ok := idx.findHashIndex(h)
	if !ok {
		return 0, plumbing.ErrObjectNotFound
	}

	offset := idx.getOffset(k, i)

	if !idx.offsetHashIsFull {
		if idx.offsetHash == nil {
			idx.offsetHash = make(map[int64]common.ObjectHash)
		}
		// Save the offset for reverse lookup
		idx.offsetHash[int64(offset)] = h
	}

	return int64(offset), nil
}

func (idx *MemoryIndex) getOffset(firstLevel, secondLevel int) uint64 {
	offset := secondLevel << 2
	ofs := encbin.BigEndian.Uint32(idx.Offset32[firstLevel][offset : offset+4])

	if (uint64(ofs) & isO64Mask) != 0 {
		offset := 8 * (uint64(ofs) & ^isO64Mask)
		n := encbin.BigEndian.Uint64(idx.Offset64[offset : offset+8])
		return n
	}

	return uint64(ofs)
}

// FindCRC32 implements the Index interface.
func (idx *MemoryIndex) FindCRC32(h common.ObjectHash) (uint32, error) {
	k := idx.FanoutMapping[h.Sum()[0]]
	i, ok := idx.findHashIndex(h)
	if !ok {
		return 0, plumbing.ErrObjectNotFound
	}

	return idx.getCRC32(k, i), nil
}

func (idx *MemoryIndex) getCRC32(firstLevel, secondLevel int) uint32 {
	offset := secondLevel << 2
	return encbin.BigEndian.Uint32(idx.CRC32[firstLevel][offset : offset+4])
}

// FindHash implements the Index interface.
func (idx *MemoryIndex) FindHash(o int64) (common.ObjectHash, error) {
	var hash common.ObjectHash
	var ok bool

	if idx.offsetHash != nil {
		if hash, ok = idx.offsetHash[o]; ok {
			return hash, nil
		}
	}

	// Lazily generate the reverse offset/hash map if required.
	if !idx.offsetHashIsFull || idx.offsetHash == nil {
		if err := idx.genOffsetHash(); err != nil {
			return idx.factory.ZeroHash(), err
		}

		hash, ok = idx.offsetHash[o]
	}

	if hash == nil {
		hash = idx.factory.ZeroHash()
	}

	if !ok {
		return hash, plumbing.ErrObjectNotFound
	}

	return hash, nil
}

// genOffsetHash generates the offset/hash mapping for reverse search.
func (idx *MemoryIndex) genOffsetHash() error {
	count, err := idx.Count()
	if err != nil {
		return err
	}

	idx.offsetHash = make(map[int64]common.ObjectHash, count)
	idx.offsetHashIsFull = true

	i := uint32(0)
	for firstLevel, fanoutValue := range idx.Fanout {
		mappedFirstLevel := idx.FanoutMapping[firstLevel]
		for secondLevel := uint32(0); i < fanoutValue; i++ {
			objectIDLength := uint32(idx.factory.Size())
			h := idx.factory.FromBytes(idx.Names[mappedFirstLevel][secondLevel*objectIDLength:])
			offset := int64(idx.getOffset(mappedFirstLevel, int(secondLevel)))
			idx.offsetHash[offset] = h
			secondLevel++
		}
	}

	return nil
}

// Count implements the Index interface.
func (idx *MemoryIndex) Count() (int64, error) {
	return int64(idx.Fanout[fanout-1]), nil
}

// Entries implements the Index interface.
func (idx *MemoryIndex) Entries() (EntryIter, error) {
	return &idxfileEntryIter{idx, 0, 0, 0}, nil
}

// EntriesByOffset implements the Index interface.
func (idx *MemoryIndex) EntriesByOffset() (EntryIter, error) {
	count, err := idx.Count()
	if err != nil {
		return nil, err
	}

	iter := &idxfileEntryOffsetIter{
		entries: make(entriesByOffset, count),
	}

	entries, err := idx.Entries()
	if err != nil {
		return nil, err
	}

	for pos := 0; int64(pos) < count; pos++ {
		entry, err := entries.Next()
		if err != nil {
			return nil, err
		}

		iter.entries[pos] = entry
	}

	sort.Sort(iter.entries)

	return iter, nil
}

// EntryIter is an iterator that will return the entries in a packfile index.
type EntryIter interface {
	// Next returns the next entry in the packfile index.
	Next() (*Entry, error)
	// Close closes the iterator.
	Close() error
}

type idxfileEntryIter struct {
	idx                     *MemoryIndex
	total                   int
	firstLevel, secondLevel int
}

func (i *idxfileEntryIter) Next() (*Entry, error) {
	for {
		if i.firstLevel >= fanout {
			return nil, io.EOF
		}

		if i.total >= int(i.idx.Fanout[i.firstLevel]) {
			i.firstLevel++
			i.secondLevel = 0
			continue
		}

		mappedFirstLevel := i.idx.FanoutMapping[i.firstLevel]
		entry := &Entry{
			Hash:   i.idx.factory.FromBytes(i.idx.Names[mappedFirstLevel][i.secondLevel*i.idx.factory.Size():]),
			Offset: i.idx.getOffset(mappedFirstLevel, i.secondLevel),
			CRC32:  i.idx.getCRC32(mappedFirstLevel, i.secondLevel),
		}

		i.secondLevel++
		i.total++

		return entry, nil
	}
}

func (i *idxfileEntryIter) Close() error {
	i.firstLevel = fanout
	return nil
}

// Entry is the in memory representation of an object entry in the idx file.
type Entry struct {
	Hash   common.ObjectHash
	CRC32  uint32
	Offset uint64
}

type idxfileEntryOffsetIter struct {
	entries entriesByOffset
	pos     int
}

func (i *idxfileEntryOffsetIter) Next() (*Entry, error) {
	if i.pos >= len(i.entries) {
		return nil, io.EOF
	}

	entry := i.entries[i.pos]
	i.pos++

	return entry, nil
}

func (i *idxfileEntryOffsetIter) Close() error {
	i.pos = len(i.entries) + 1
	return nil
}

type entriesByOffset []*Entry

func (o entriesByOffset) Len() int {
	return len(o)
}

func (o entriesByOffset) Less(i int, j int) bool {
	return o[i].Offset < o[j].Offset
}

func (o entriesByOffset) Swap(i int, j int) {
	o[i], o[j] = o[j], o[i]
}
