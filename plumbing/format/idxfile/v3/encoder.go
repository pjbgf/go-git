package idxfile

import (
	"fmt"
	"io"

	format "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/go-git/go-git/v5/utils/binary"
)

// Encoder writes MemoryIndex structs to an output stream.
type Encoder struct {
	io.Writer
	hash         hash.Hash
	objectFormat format.ObjectFormat
	version      Version
}

type EncoderOptions struct {
	format.ObjectFormat
	Version
}

// NewEncoder returns a new stream encoder that writes to w.
func NewEncoder(w io.Writer) (*Encoder, error) {
	// TODO: Review - may want to rely on the built type instead
	return NewEncoderWithOptions(w, EncoderOptions{ObjectFormat: format.SHA1})
}

func NewEncoderWithOptions(w io.Writer, opts EncoderOptions) (*Encoder, error) {
	if w == nil {
		return nil, fmt.Errorf("cannot create encoder: nil writer")
	}

	h, err := hash.FromObjectFormat(opts.ObjectFormat)
	if err != nil {
		return nil, err
	}

	mw := io.MultiWriter(w, h)

	v := Version2
	if opts.Version == Version3 {
		v = Version3
	}

	return &Encoder{mw, h, opts.ObjectFormat, v}, nil
}

// Encode encodes an MemoryIndex to the encoder writer.
func (e *Encoder) Encode(idx *MemoryIndex) (int, error) {
	if idx == nil {
		return 0, fmt.Errorf("failed to encode: target index is nil")
	}

	if !idx.m.TryLock() {
		return 0, fmt.Errorf("failed to encode: %w", ErrMemoryIndexLocked)
	}
	defer idx.m.Unlock()

	encodeFlow := []func(*MemoryIndex) (int, error){
		e.encodeHeader,
		e.encodeFanout,
		e.encodeHashes,
		e.encodeCRC32,
		e.encodeOffsets,
		e.encodeChecksums,
	}

	sz := 0
	for _, f := range encodeFlow {
		i, err := f(idx)
		sz += i

		if err != nil {
			return sz, err
		}
	}

	return sz, nil
}

func (e *Encoder) encodeHeader(idx *MemoryIndex) (int, error) {
	c, err := e.Write(idxHeader)
	if err != nil {
		return c, err
	}

	return c + 4, binary.WriteUint32(e, uint32(e.version))
}

func (e *Encoder) encodeFanout(idx *MemoryIndex) (int, error) {
	for _, c := range idx.Fanout {
		if err := binary.WriteUint32(e, c); err != nil {
			return 0, err
		}
	}

	return fanout * 4, nil
}

func (e *Encoder) encodeHashes(idx *MemoryIndex) (int, error) {
	var size int
	for k := 0; k < fanout; k++ {
		pos := idx.FanoutMapping[k]
		if pos == noMapping {
			continue
		}

		n, err := e.Write(idx.Names[pos])
		if err != nil {
			return size, err
		}
		size += n
	}
	return size, nil
}

func (e *Encoder) encodeCRC32(idx *MemoryIndex) (int, error) {
	var size int
	for k := 0; k < fanout; k++ {
		pos := idx.FanoutMapping[k]
		if pos == noMapping {
			continue
		}

		n, err := e.Write(idx.CRC32[pos])
		if err != nil {
			return size, err
		}

		size += n
	}

	return size, nil
}

func (e *Encoder) encodeOffsets(idx *MemoryIndex) (int, error) {
	var size int
	for k := 0; k < fanout; k++ {
		pos := idx.FanoutMapping[k]
		if pos == noMapping {
			continue
		}

		n, err := e.Write(idx.Offset32[pos])
		if err != nil {
			return size, err
		}

		size += n
	}

	if len(idx.Offset64) > 0 {
		n, err := e.Write(idx.Offset64)
		if err != nil {
			return size, err
		}

		size += n
	}

	return size, nil
}

func (e *Encoder) encodeChecksums(idx *MemoryIndex) (int, error) {
	if _, err := e.Write(idx.PackfileChecksum[:]); err != nil {
		return 0, err
	}

	copy(idx.IdxChecksum[:], e.hash.Sum(nil)[:e.hash.Size()])
	if _, err := e.Write(idx.IdxChecksum[:]); err != nil {
		return 0, err
	}

	return e.hash.Size() * 2, nil
}
