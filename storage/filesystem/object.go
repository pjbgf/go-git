package filesystem

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/utils/ioutil"

	"github.com/go-git/go-billy/v5"
)

type ObjectStorage struct {
	options Options

	// objectCache is an object cache uses to cache delta's bases and also recently
	// loaded loose objects
	objectCache cache.Object

	dir   *dotgit.DotGit
	index map[common.ObjectHash]idxfile.Index

	packList    []common.ObjectHash
	packListIdx int
	packfiles   map[common.ObjectHash]*packfile.Packfile
	factory     common.HashFactory
}

// NewObjectStorage creates a new ObjectStorage with the given .git directory and cache.
func NewObjectStorage(dir *dotgit.DotGit, objectCache cache.Object) *ObjectStorage {
	return NewObjectStorageWithOptions(dir, objectCache, Options{})
}

// NewObjectStorageWithOptions creates a new ObjectStorage with the given .git directory, cache and extra options
func NewObjectStorageWithOptions(dir *dotgit.DotGit, objectCache cache.Object, ops Options) *ObjectStorage {
	return &ObjectStorage{
		options:     ops,
		objectCache: objectCache,
		dir:         dir,
		// TODO: fix DIP violation as part of sha256
		factory: sha1.Factory,
	}
}

func (s *ObjectStorage) requireIndex() error {
	if s.index != nil {
		return nil
	}

	s.index = make(map[common.ObjectHash]idxfile.Index)
	packs, err := s.dir.ObjectPacks()
	if err != nil {
		return err
	}

	for _, h := range packs {
		if err := s.loadIdxFile(h); err != nil {
			return err
		}
	}

	return nil
}

// Reindex indexes again all packfiles. Useful if git changed packfiles externally
func (s *ObjectStorage) Reindex() {
	s.index = nil
}

func (s *ObjectStorage) loadIdxFile(h common.ObjectHash) (err error) {
	f, err := s.dir.ObjectPackIdx(h)
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(f, &err)

	idxf := idxfile.NewMemoryIndex(s.factory)
	d, err := idxfile.NewDecoder(f, s.factory)
	if err != nil {
		return err
	}

	err = d.Decode(idxf)
	if err != nil {
		return err
	}

	s.index[h] = idxf
	return err
}

func (s *ObjectStorage) NewEncodedObject() plumbing.EncodedObject {
	return &plumbing.MemoryObject{}
}

func (s *ObjectStorage) PackfileWriter() (io.WriteCloser, error) {
	if err := s.requireIndex(); err != nil {
		return nil, err
	}

	w, err := s.dir.NewObjectPack()
	if err != nil {
		return nil, err
	}

	w.Notify = func(h common.ObjectHash, writer *idxfile.Writer) {
		index, err := writer.Index()
		if err == nil {
			s.index[h] = index
		}
	}

	return w, nil
}

// SetEncodedObject adds a new object to the storage.
func (s *ObjectStorage) SetEncodedObject(o plumbing.EncodedObject) (h common.ObjectHash, err error) {
	if o.Type() == plumbing.OFSDeltaObject || o.Type() == plumbing.REFDeltaObject {
		return sha1.ZeroHash(), plumbing.ErrInvalidType
	}

	ow, err := s.dir.NewObject()
	if err != nil {
		return sha1.ZeroHash(), err
	}

	defer ioutil.CheckClose(ow, &err)

	or, err := o.Reader()
	if err != nil {
		return sha1.ZeroHash(), err
	}

	defer ioutil.CheckClose(or, &err)

	if err = ow.WriteHeader(o.Type(), o.Size()); err != nil {
		return sha1.ZeroHash(), err
	}

	if _, err = io.Copy(ow, or); err != nil {
		return sha1.ZeroHash(), err
	}

	return o.Hash(), err
}

// LazyWriter returns a lazy ObjectWriter that is bound to a DotGit file.
// It first write the header passing on the object type and size, so
// that the object contents can be written later, without the need to
// create a MemoryObject and buffering its entire contents into memory.
func (s *ObjectStorage) LazyWriter() (w io.WriteCloser, wh func(typ plumbing.ObjectType, sz int64) error, err error) {
	ow, err := s.dir.NewObject()
	if err != nil {
		return nil, nil, err
	}

	return ow, ow.WriteHeader, nil
}

// HasEncodedObject returns nil if the object exists, without actually
// reading the object data from storage.
func (s *ObjectStorage) HasEncodedObject(h common.ObjectHash) (err error) {
	// Check unpacked objects
	f, err := s.dir.Object(h)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// Fall through to check packed objects.
	} else {
		defer ioutil.CheckClose(f, &err)
		return nil
	}

	// Check packed objects.
	if err := s.requireIndex(); err != nil {
		return err
	}
	_, _, offset := s.findObjectInPackfile(h)
	if offset == -1 {
		return plumbing.ErrObjectNotFound
	}
	return nil
}

func (s *ObjectStorage) encodedObjectSizeFromUnpacked(h common.ObjectHash) (
	size int64, err error) {
	f, err := s.dir.Object(h)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, plumbing.ErrObjectNotFound
		}

		return 0, err
	}

	r, err := objfile.NewReader(f)
	if err != nil {
		return 0, err
	}
	defer ioutil.CheckClose(r, &err)

	_, size, err = r.Header()
	return size, err
}

func (s *ObjectStorage) packfile(idx idxfile.Index, pack common.ObjectHash) (*packfile.Packfile, error) {
	if p := s.packfileFromCache(pack); p != nil {
		return p, nil
	}

	f, err := s.dir.ObjectPack(pack)
	if err != nil {
		return nil, err
	}

	var p *packfile.Packfile
	if s.objectCache != nil {
		p = packfile.NewPackfileWithCache(idx, s.dir.Fs(), f, s.objectCache, s.options.LargeObjectThreshold)
	} else {
		p = packfile.NewPackfile(idx, s.dir.Fs(), f, s.options.LargeObjectThreshold)
	}

	return p, s.storePackfileInCache(pack, p)
}

func (s *ObjectStorage) packfileFromCache(hash common.ObjectHash) *packfile.Packfile {
	if s.packfiles == nil {
		if s.options.KeepDescriptors {
			s.packfiles = make(map[common.ObjectHash]*packfile.Packfile)
		} else if s.options.MaxOpenDescriptors > 0 {
			s.packList = make([]common.ObjectHash, s.options.MaxOpenDescriptors)
			s.packfiles = make(map[common.ObjectHash]*packfile.Packfile, s.options.MaxOpenDescriptors)
		}
	}

	return s.packfiles[hash]
}

func (s *ObjectStorage) storePackfileInCache(hash common.ObjectHash, p *packfile.Packfile) error {
	if s.options.KeepDescriptors {
		s.packfiles[hash] = p
		return nil
	}

	if s.options.MaxOpenDescriptors <= 0 {
		return nil
	}

	// start over as the limit of packList is hit
	if s.packListIdx >= len(s.packList) {
		s.packListIdx = 0
	}

	// close the existing packfile if open
	if next := s.packList[s.packListIdx]; next == nil || !next.IsZero() {
		open := s.packfiles[next]
		delete(s.packfiles, next)
		if open != nil {
			if err := open.Close(); err != nil {
				return err
			}
		}
	}

	// cache newly open packfile
	s.packList[s.packListIdx] = hash
	s.packfiles[hash] = p
	s.packListIdx++

	return nil
}

func (s *ObjectStorage) encodedObjectSizeFromPackfile(h common.ObjectHash) (
	size int64, err error) {
	if err := s.requireIndex(); err != nil {
		return 0, err
	}

	pack, _, offset := s.findObjectInPackfile(h)
	if offset == -1 {
		return 0, plumbing.ErrObjectNotFound
	}

	idx := s.index[pack]
	hash, err := idx.FindHash(offset)
	if err == nil {
		obj, ok := s.objectCache.Get(hash)
		if ok {
			return obj.Size(), nil
		}
	} else if err != nil && err != plumbing.ErrObjectNotFound {
		return 0, err
	}

	p, err := s.packfile(idx, pack)
	if err != nil {
		return 0, err
	}

	if !s.options.KeepDescriptors && s.options.MaxOpenDescriptors == 0 {
		defer ioutil.CheckClose(p, &err)
	}

	return p.GetSizeByOffset(offset)
}

// EncodedObjectSize returns the plaintext size of the given object,
// without actually reading the full object data from storage.
func (s *ObjectStorage) EncodedObjectSize(h common.ObjectHash) (
	size int64, err error) {
	size, err = s.encodedObjectSizeFromUnpacked(h)
	if err != nil && err != plumbing.ErrObjectNotFound {
		return 0, err
	} else if err == nil {
		return size, nil
	}

	return s.encodedObjectSizeFromPackfile(h)
}

// EncodedObject returns the object with the given hash, by searching for it in
// the packfile and the git object directories.
func (s *ObjectStorage) EncodedObject(t plumbing.ObjectType, h common.ObjectHash) (plumbing.EncodedObject, error) {
	var obj plumbing.EncodedObject
	var err error

	if s.index != nil {
		obj, err = s.getFromPackfile(h, false)
		if err == plumbing.ErrObjectNotFound {
			obj, err = s.getFromUnpacked(h)
		}
	} else {
		obj, err = s.getFromUnpacked(h)
		if err == plumbing.ErrObjectNotFound {
			obj, err = s.getFromPackfile(h, false)
		}
	}

	// If the error is still object not found, check if it's a shared object
	// repository.
	if err == plumbing.ErrObjectNotFound {
		dotgits, e := s.dir.Alternates()
		if e == nil {
			// Create a new object storage with the DotGit(s) and check for the
			// required hash object. Skip when not found.
			for _, dg := range dotgits {
				o := NewObjectStorage(dg, s.objectCache)
				enobj, enerr := o.EncodedObject(t, h)
				if enerr != nil {
					continue
				}
				return enobj, nil
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if plumbing.AnyObject != t && obj.Type() != t {
		return nil, plumbing.ErrObjectNotFound
	}

	return obj, nil
}

// DeltaObject returns the object with the given hash, by searching for
// it in the packfile and the git object directories.
func (s *ObjectStorage) DeltaObject(t plumbing.ObjectType,
	h common.ObjectHash) (plumbing.EncodedObject, error) {
	obj, err := s.getFromUnpacked(h)
	if err == plumbing.ErrObjectNotFound {
		obj, err = s.getFromPackfile(h, true)
	}

	if err != nil {
		return nil, err
	}

	if plumbing.AnyObject != t && obj.Type() != t {
		return nil, plumbing.ErrObjectNotFound
	}

	return obj, nil
}

func (s *ObjectStorage) getFromUnpacked(h common.ObjectHash) (obj plumbing.EncodedObject, err error) {
	f, err := s.dir.Object(h)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, plumbing.ErrObjectNotFound
		}

		return nil, err
	}
	defer ioutil.CheckClose(f, &err)

	if cacheObj, found := s.objectCache.Get(h); found {
		return cacheObj, nil
	}

	r, err := objfile.NewReader(f)
	if err != nil {
		return nil, err
	}

	defer ioutil.CheckClose(r, &err)

	t, size, err := r.Header()
	if err != nil {
		return nil, err
	}

	if s.options.LargeObjectThreshold > 0 && size > s.options.LargeObjectThreshold {
		obj = dotgit.NewEncodedObject(s.dir, h, t, size)
		return obj, nil
	}

	obj = s.NewEncodedObject()

	obj.SetType(t)
	obj.SetSize(size)
	w, err := obj.Writer()
	if err != nil {
		return nil, err
	}

	defer ioutil.CheckClose(w, &err)

	s.objectCache.Put(obj)

	bufp := copyBufferPool.Get().(*[]byte)
	buf := *bufp
	_, err = io.CopyBuffer(w, r, buf)
	copyBufferPool.Put(bufp)

	return obj, err
}

var copyBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return &b
	},
}

// Get returns the object with the given hash, by searching for it in
// the packfile.
func (s *ObjectStorage) getFromPackfile(h common.ObjectHash, canBeDelta bool) (
	plumbing.EncodedObject, error) {

	if err := s.requireIndex(); err != nil {
		return nil, err
	}

	pack, hash, offset := s.findObjectInPackfile(h)
	if offset == -1 {
		return nil, plumbing.ErrObjectNotFound
	}

	idx := s.index[pack]
	p, err := s.packfile(idx, pack)
	if err != nil {
		return nil, err
	}

	if !s.options.KeepDescriptors && s.options.MaxOpenDescriptors == 0 {
		defer ioutil.CheckClose(p, &err)
	}

	if canBeDelta {
		return s.decodeDeltaObjectAt(p, offset, hash)
	}

	return s.decodeObjectAt(p, offset)
}

func (s *ObjectStorage) decodeObjectAt(
	p *packfile.Packfile,
	offset int64,
) (plumbing.EncodedObject, error) {
	hash, err := p.FindHash(offset)
	if err == nil {
		obj, ok := s.objectCache.Get(hash)
		if ok {
			return obj, nil
		}
	}

	if err != nil && err != plumbing.ErrObjectNotFound {
		return nil, err
	}

	return p.GetByOffset(offset)
}

func (s *ObjectStorage) decodeDeltaObjectAt(
	p *packfile.Packfile,
	offset int64,
	hash common.ObjectHash,
) (plumbing.EncodedObject, error) {
	scan := p.Scanner()
	header, err := scan.SeekObjectHeader(offset)
	if err != nil {
		return nil, err
	}

	var (
		base common.ObjectHash
	)

	switch header.Type {
	case plumbing.REFDeltaObject:
		base = header.Reference
	case plumbing.OFSDeltaObject:
		base, err = p.FindHash(header.OffsetReference)
		if err != nil {
			return nil, err
		}
	default:
		return s.decodeObjectAt(p, offset)
	}

	obj := &plumbing.MemoryObject{}
	obj.SetType(header.Type)
	w, err := obj.Writer()
	if err != nil {
		return nil, err
	}

	if _, _, err := scan.NextObject(w); err != nil {
		return nil, err
	}

	return newDeltaObject(obj, hash, base, header.Length), nil
}

func (s *ObjectStorage) findObjectInPackfile(h common.ObjectHash) (common.ObjectHash, common.ObjectHash, int64) {
	for packfile, index := range s.index {
		offset, err := index.FindOffset(h)
		if err == nil {
			return packfile, h, offset
		}
	}

	return sha1.ZeroHash(), sha1.ZeroHash(), -1
}

// HashesWithPrefix returns all objects with a hash that starts with a prefix by searching for
// them in the packfile and the git object directories.
func (s *ObjectStorage) HashesWithPrefix(prefix []byte) ([]common.ObjectHash, error) {
	hashes, err := s.dir.ObjectsWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	seen := hashListAsMap(hashes)

	// TODO: This could be faster with some idxfile changes,
	// or diving into the packfile.
	if err := s.requireIndex(); err != nil {
		return nil, err
	}
	for _, index := range s.index {
		ei, err := index.Entries()
		if err != nil {
			return nil, err
		}
		for {
			e, err := ei.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			if bytes.HasPrefix(e.Hash.Sum(), prefix) {
				if _, ok := seen[e.Hash]; ok {
					continue
				}
				hashes = append(hashes, e.Hash)
			}
		}
		ei.Close()
	}

	return hashes, nil
}

// IterEncodedObjects returns an iterator for all the objects in the packfile
// with the given type.
func (s *ObjectStorage) IterEncodedObjects(t plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	objects, err := s.dir.Objects()
	if err != nil {
		return nil, err
	}

	seen := make(map[common.ObjectHash]struct{})
	var iters []storer.EncodedObjectIter
	if len(objects) != 0 {
		iters = append(iters, &objectsIter{s: s, t: t, h: objects})
		seen = hashListAsMap(objects)
	}

	packi, err := s.buildPackfileIters(t, seen)
	if err != nil {
		return nil, err
	}

	iters = append(iters, packi)
	return storer.NewMultiEncodedObjectIter(iters), nil
}

func (s *ObjectStorage) buildPackfileIters(
	t plumbing.ObjectType,
	seen map[common.ObjectHash]struct{},
) (storer.EncodedObjectIter, error) {
	if err := s.requireIndex(); err != nil {
		return nil, err
	}

	packs, err := s.dir.ObjectPacks()
	if err != nil {
		return nil, err
	}
	return &lazyPackfilesIter{
		hashes: packs,
		open: func(h common.ObjectHash) (storer.EncodedObjectIter, error) {
			pack, err := s.dir.ObjectPack(h)
			if err != nil {
				return nil, err
			}
			return newPackfileIter(
				s.dir.Fs(), pack, t, seen, s.index[h],
				s.objectCache, s.options.KeepDescriptors,
				s.options.LargeObjectThreshold,
			)
		},
	}, nil
}

// Close closes all opened files.
func (s *ObjectStorage) Close() error {
	var firstError error
	if s.options.KeepDescriptors || s.options.MaxOpenDescriptors > 0 {
		for _, packfile := range s.packfiles {
			err := packfile.Close()
			if firstError == nil && err != nil {
				firstError = err
			}
		}
	}

	s.packfiles = nil
	s.dir.Close()

	return firstError
}

type lazyPackfilesIter struct {
	hashes []common.ObjectHash
	open   func(h common.ObjectHash) (storer.EncodedObjectIter, error)
	cur    storer.EncodedObjectIter
}

func (it *lazyPackfilesIter) Next() (plumbing.EncodedObject, error) {
	for {
		if it.cur == nil {
			if len(it.hashes) == 0 {
				return nil, io.EOF
			}
			h := it.hashes[0]
			it.hashes = it.hashes[1:]

			sub, err := it.open(h)
			if err == io.EOF {
				continue
			} else if err != nil {
				return nil, err
			}
			it.cur = sub
		}
		ob, err := it.cur.Next()
		if err == io.EOF {
			it.cur.Close()
			it.cur = nil
			continue
		} else if err != nil {
			return nil, err
		}
		return ob, nil
	}
}

func (it *lazyPackfilesIter) ForEach(cb func(plumbing.EncodedObject) error) error {
	return storer.ForEachIterator(it, cb)
}

func (it *lazyPackfilesIter) Close() {
	if it.cur != nil {
		it.cur.Close()
		it.cur = nil
	}
	it.hashes = nil
}

type packfileIter struct {
	pack billy.File
	iter storer.EncodedObjectIter
	seen map[common.ObjectHash]struct{}

	// tells whether the pack file should be left open after iteration or not
	keepPack bool
}

// NewPackfileIter returns a new EncodedObjectIter for the provided packfile
// and object type. Packfile and index file will be closed after they're
// used. If keepPack is true the packfile won't be closed after the iteration
// finished.
func NewPackfileIter(
	fs billy.Filesystem,
	f billy.File,
	idxFile billy.File,
	t plumbing.ObjectType,
	keepPack bool,
	largeObjectThreshold int64,
) (storer.EncodedObjectIter, error) {
	// TODO: fix DIP violation as part of sha256
	idx := idxfile.NewMemoryIndex(sha1.Factory)
	d, err := idxfile.NewDecoder(idxFile, sha1.Factory)
	if err != nil {
		return nil, err
	}

	err = d.Decode(idx)
	if err != nil {
		return nil, err
	}

	if err := idxFile.Close(); err != nil {
		return nil, err
	}

	seen := make(map[common.ObjectHash]struct{})
	return newPackfileIter(fs, f, t, seen, idx, nil, keepPack, largeObjectThreshold)
}

func newPackfileIter(
	fs billy.Filesystem,
	f billy.File,
	t plumbing.ObjectType,
	seen map[common.ObjectHash]struct{},
	index idxfile.Index,
	cache cache.Object,
	keepPack bool,
	largeObjectThreshold int64,
) (storer.EncodedObjectIter, error) {
	var p *packfile.Packfile
	if cache != nil {
		p = packfile.NewPackfileWithCache(index, fs, f, cache, largeObjectThreshold)
	} else {
		p = packfile.NewPackfile(index, fs, f, largeObjectThreshold)
	}

	iter, err := p.GetByType(t)
	if err != nil {
		return nil, err
	}

	return &packfileIter{
		pack:     f,
		iter:     iter,
		seen:     seen,
		keepPack: keepPack,
	}, nil
}

func (iter *packfileIter) Next() (plumbing.EncodedObject, error) {
	for {
		obj, err := iter.iter.Next()
		if err != nil {
			return nil, err
		}

		if _, ok := iter.seen[obj.Hash()]; ok {
			continue
		}

		return obj, nil
	}
}

func (iter *packfileIter) ForEach(cb func(plumbing.EncodedObject) error) error {
	for {
		o, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				iter.Close()
				return nil
			}
			return err
		}

		if err := cb(o); err != nil {
			return err
		}
	}
}

func (iter *packfileIter) Close() {
	iter.iter.Close()
	if !iter.keepPack {
		_ = iter.pack.Close()
	}
}

type objectsIter struct {
	s *ObjectStorage
	t plumbing.ObjectType
	h []common.ObjectHash
}

func (iter *objectsIter) Next() (plumbing.EncodedObject, error) {
	if len(iter.h) == 0 {
		return nil, io.EOF
	}

	obj, err := iter.s.getFromUnpacked(iter.h[0])
	iter.h = iter.h[1:]

	if err != nil {
		return nil, err
	}

	if iter.t != plumbing.AnyObject && iter.t != obj.Type() {
		return iter.Next()
	}

	return obj, err
}

func (iter *objectsIter) ForEach(cb func(plumbing.EncodedObject) error) error {
	for {
		o, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := cb(o); err != nil {
			return err
		}
	}
}

func (iter *objectsIter) Close() {
	iter.h = []common.ObjectHash{}
}

func hashListAsMap(l []common.ObjectHash) map[common.ObjectHash]struct{} {
	m := make(map[common.ObjectHash]struct{}, len(l))
	for _, h := range l {
		m[h] = struct{}{}
	}
	return m
}

func (s *ObjectStorage) ForEachObjectHash(fun func(common.ObjectHash) error) error {
	err := s.dir.ForEachObjectHash(fun)
	if err == storer.ErrStop {
		return nil
	}
	return err
}

func (s *ObjectStorage) LooseObjectTime(hash common.ObjectHash) (time.Time, error) {
	fi, err := s.dir.ObjectStat(hash)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

func (s *ObjectStorage) DeleteLooseObject(hash common.ObjectHash) error {
	return s.dir.ObjectDelete(hash)
}

func (s *ObjectStorage) ObjectPacks() ([]common.ObjectHash, error) {
	return s.dir.ObjectPacks()
}

func (s *ObjectStorage) DeleteOldObjectPackAndIndex(h common.ObjectHash, t time.Time) error {
	return s.dir.DeleteOldObjectPackAndIndex(h, t)
}
