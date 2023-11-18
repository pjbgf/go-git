package packfile

import (
	"compress/zlib"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/utils/binary"
	"github.com/go-git/go-git/v5/utils/ioutil"
)

// Encoder gets the data from the storage and write it into the writer in PACK
// format
type Encoder struct {
	selector *deltaSelector
	w        *offsetWriter
	zw       *zlib.Writer
	hasher   hash.Hash
	factory  common.HashFactory

	useRefDeltas bool
}

// NewEncoder creates a new packfile encoder using a specific Writer and
// EncodedObjectStorer. By default deltas used to generate the packfile will be
// OFSDeltaObject. To use Reference deltas, set useRefDeltas to true.
func NewEncoder(w io.Writer, s storer.EncodedObjectStorer, useRefDeltas bool, h hash.Hash, f common.HashFactory) *Encoder {
	mw := io.MultiWriter(w, h)
	ow := newOffsetWriter(mw)
	zw := zlib.NewWriter(mw)
	return &Encoder{
		selector:     newDeltaSelector(s),
		w:            ow,
		zw:           zw,
		hasher:       h,
		useRefDeltas: useRefDeltas,
		factory:      f,
	}
}

// Encode creates a packfile containing all the objects referenced in
// hashes and writes it to the writer in the Encoder.  `packWindow`
// specifies the size of the sliding window used to compare objects
// for delta compression; 0 turns off delta compression entirely.
func (e *Encoder) Encode(
	hashes []common.ObjectHash,
	packWindow uint,
) (common.ObjectHash, error) {
	objects, err := e.selector.ObjectsToPack(hashes, packWindow)
	if err != nil {
		return sha1.ZeroHash(), err
	}

	return e.encode(objects)
}

func (e *Encoder) encode(objects []*ObjectToPack) (common.ObjectHash, error) {
	if err := e.head(len(objects)); err != nil {
		return sha1.ZeroHash(), err
	}

	for _, o := range objects {
		if err := e.entry(o); err != nil {
			return sha1.ZeroHash(), err
		}
	}

	return e.footer()
}

func (e *Encoder) head(numEntries int) error {
	return binary.Write(
		e.w,
		signature,
		int32(VersionSupported),
		int32(numEntries),
	)
}

func (e *Encoder) entry(o *ObjectToPack) (err error) {
	if o.WantWrite() {
		// A cycle exists in this delta chain. This should only occur if a
		// selected object representation disappeared during writing
		// (for example due to a concurrent repack) and a different base
		// was chosen, forcing a cycle. Select something other than a
		// delta, and write this object.
		e.selector.restoreOriginal(o)
		o.BackToOriginal()
	}

	if o.IsWritten() {
		return nil
	}

	o.MarkWantWrite()

	if err := e.writeBaseIfDelta(o); err != nil {
		return err
	}

	// We need to check if we already write that object due a cyclic delta chain
	if o.IsWritten() {
		return nil
	}

	o.Offset = e.w.Offset()

	if o.IsDelta() {
		if err := e.writeDeltaHeader(o); err != nil {
			return err
		}
	} else {
		if err := e.entryHead(o.Type(), o.Size()); err != nil {
			return err
		}
	}

	e.zw.Reset(e.w)

	defer ioutil.CheckClose(e.zw, &err)

	or, err := o.Object.Reader()
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(or, &err)

	_, err = io.Copy(e.zw, or)
	return err
}

func (e *Encoder) writeBaseIfDelta(o *ObjectToPack) error {
	if o.IsDelta() && !o.Base.IsWritten() {
		// We must write base first
		return e.entry(o.Base)
	}

	return nil
}

func (e *Encoder) writeDeltaHeader(o *ObjectToPack) error {
	// Write offset deltas by default
	t := plumbing.OFSDeltaObject
	if e.useRefDeltas {
		t = plumbing.REFDeltaObject
	}

	if err := e.entryHead(t, o.Object.Size()); err != nil {
		return err
	}

	if e.useRefDeltas {
		return e.writeRefDeltaHeader(o.Base.Hash())
	} else {
		return e.writeOfsDeltaHeader(o)
	}
}

func (e *Encoder) writeRefDeltaHeader(base common.ObjectHash) error {
	return binary.Write(e.w, base)
}

func (e *Encoder) writeOfsDeltaHeader(o *ObjectToPack) error {
	// for OFS_DELTA, offset of the base is interpreted as negative offset
	// relative to the type-byte of the header of the ofs-delta entry.
	relativeOffset := o.Offset - o.Base.Offset
	if relativeOffset <= 0 {
		return fmt.Errorf("bad offset for OFS_DELTA entry: %d", relativeOffset)
	}

	return binary.WriteVariableWidthInt(e.w, relativeOffset)
}

func (e *Encoder) entryHead(typeNum plumbing.ObjectType, size int64) error {
	t := int64(typeNum)
	header := []byte{}
	c := (t << firstLengthBits) | (size & maskFirstLength)
	size >>= firstLengthBits
	for {
		if size == 0 {
			break
		}
		header = append(header, byte(c|maskContinue))
		c = size & int64(maskLength)
		size >>= lengthBits
	}

	header = append(header, byte(c))
	_, err := e.w.Write(header)

	return err
}

func (e *Encoder) footer() (common.ObjectHash, error) {
	h := e.factory.FromBytes(e.hasher.Sum(nil))
	return h, binary.Write(e.w, h)
}

type offsetWriter struct {
	w      io.Writer
	offset int64
}

func newOffsetWriter(w io.Writer) *offsetWriter {
	return &offsetWriter{w: w}
}

func (ow *offsetWriter) Write(p []byte) (n int, err error) {
	n, err = ow.w.Write(p)
	ow.offset += int64(n)
	return n, err
}

func (ow *offsetWriter) Offset() int64 {
	return ow.offset
}
