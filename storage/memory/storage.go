// Package memory is a storage backend base on memory
package memory

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/hash/common"
	"github.com/go-git/go-git/v5/plumbing/hash/sha1"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
)

var ErrUnsupportedObjectType = fmt.Errorf("unsupported object type")

// Storage is an implementation of git.Storer that stores data on memory, being
// ephemeral. The use of this storage should be done in controlled environments,
// since the representation in memory of some repository can fill the machine
// memory. in the other hand this storage has the best performance.
type Storage struct {
	ConfigStorage
	ObjectStorage
	ShallowStorage
	IndexStorage
	ReferenceStorage
	ModuleStorage
	factory common.HashFactory
}

// NewStorage returns a new Storage base on memory
func NewStorage() *Storage {
	return &Storage{
		ReferenceStorage: make(ReferenceStorage),
		ConfigStorage:    ConfigStorage{},
		ShallowStorage:   ShallowStorage{},
		ObjectStorage: ObjectStorage{
			Objects: make(map[common.ObjectHash]plumbing.EncodedObject),
			Commits: make(map[common.ObjectHash]plumbing.EncodedObject),
			Trees:   make(map[common.ObjectHash]plumbing.EncodedObject),
			Blobs:   make(map[common.ObjectHash]plumbing.EncodedObject),
			Tags:    make(map[common.ObjectHash]plumbing.EncodedObject),
		},
		ModuleStorage: make(ModuleStorage),
		factory:       sha1.Factory,
	}
}

type ConfigStorage struct {
	config *config.Config
}

func (c *ConfigStorage) SetConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	c.config = cfg
	return nil
}

func (c *ConfigStorage) Config() (*config.Config, error) {
	if c.config == nil {
		c.config = config.NewConfig()
	}

	return c.config, nil
}

type IndexStorage struct {
	index *index.Index
}

func (c *IndexStorage) SetIndex(idx *index.Index) error {
	c.index = idx
	return nil
}

func (c *IndexStorage) Index() (*index.Index, error) {
	if c.index == nil {
		c.index = &index.Index{Version: 2}
	}

	return c.index, nil
}

type ObjectStorage struct {
	Objects map[common.ObjectHash]plumbing.EncodedObject
	Commits map[common.ObjectHash]plumbing.EncodedObject
	Trees   map[common.ObjectHash]plumbing.EncodedObject
	Blobs   map[common.ObjectHash]plumbing.EncodedObject
	Tags    map[common.ObjectHash]plumbing.EncodedObject
	factory common.HashFactory
}

func (o *ObjectStorage) NewEncodedObject() plumbing.EncodedObject {
	return &plumbing.MemoryObject{}
}

func (o *ObjectStorage) SetEncodedObject(obj plumbing.EncodedObject) (common.ObjectHash, error) {
	h := obj.Hash()
	o.Objects[h] = obj

	switch obj.Type() {
	case plumbing.CommitObject:
		o.Commits[h] = o.Objects[h]
	case plumbing.TreeObject:
		o.Trees[h] = o.Objects[h]
	case plumbing.BlobObject:
		o.Blobs[h] = o.Objects[h]
	case plumbing.TagObject:
		o.Tags[h] = o.Objects[h]
	default:
		return h, ErrUnsupportedObjectType
	}

	return h, nil
}

func (o *ObjectStorage) HasEncodedObject(h common.ObjectHash) (err error) {
	if _, ok := o.Objects[h]; !ok {
		return plumbing.ErrObjectNotFound
	}
	return nil
}

func (o *ObjectStorage) EncodedObjectSize(h common.ObjectHash) (
	size int64, err error) {
	obj, ok := o.Objects[h]
	if !ok {
		return 0, plumbing.ErrObjectNotFound
	}

	return obj.Size(), nil
}

func (o *ObjectStorage) EncodedObject(t plumbing.ObjectType, h common.ObjectHash) (plumbing.EncodedObject, error) {
	obj, ok := o.Objects[h]
	if !ok || (plumbing.AnyObject != t && obj.Type() != t) {
		return nil, plumbing.ErrObjectNotFound
	}

	return obj, nil
}

func (o *ObjectStorage) IterEncodedObjects(t plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	var series []plumbing.EncodedObject
	switch t {
	case plumbing.AnyObject:
		series = flattenObjectMap(o.Objects)
	case plumbing.CommitObject:
		series = flattenObjectMap(o.Commits)
	case plumbing.TreeObject:
		series = flattenObjectMap(o.Trees)
	case plumbing.BlobObject:
		series = flattenObjectMap(o.Blobs)
	case plumbing.TagObject:
		series = flattenObjectMap(o.Tags)
	}

	return storer.NewEncodedObjectSliceIter(series), nil
}

func flattenObjectMap(m map[common.ObjectHash]plumbing.EncodedObject) []plumbing.EncodedObject {
	objects := make([]plumbing.EncodedObject, 0, len(m))
	for _, obj := range m {
		objects = append(objects, obj)
	}
	return objects
}

func (o *ObjectStorage) Begin() storer.Transaction {
	return &TxObjectStorage{
		Storage: o,
		Objects: make(map[common.ObjectHash]plumbing.EncodedObject),
	}
}

func (o *ObjectStorage) ForEachObjectHash(fun func(common.ObjectHash) error) error {
	for h := range o.Objects {
		err := fun(h)
		if err != nil {
			if err == storer.ErrStop {
				return nil
			}
			return err
		}
	}
	return nil
}

func (o *ObjectStorage) ObjectPacks() ([]common.ObjectHash, error) {
	return nil, nil
}
func (o *ObjectStorage) DeleteOldObjectPackAndIndex(common.ObjectHash, time.Time) error {
	return nil
}

var errNotSupported = fmt.Errorf("not supported")

func (o *ObjectStorage) LooseObjectTime(hash common.ObjectHash) (time.Time, error) {
	return time.Time{}, errNotSupported
}
func (o *ObjectStorage) DeleteLooseObject(common.ObjectHash) error {
	return errNotSupported
}

func (o *ObjectStorage) AddAlternate(remote string) error {
	return errNotSupported
}

type TxObjectStorage struct {
	Storage *ObjectStorage
	Objects map[common.ObjectHash]plumbing.EncodedObject
}

func (tx *TxObjectStorage) SetEncodedObject(obj plumbing.EncodedObject) (common.ObjectHash, error) {
	h := obj.Hash()
	tx.Objects[h] = obj

	return h, nil
}

func (tx *TxObjectStorage) EncodedObject(t plumbing.ObjectType, h common.ObjectHash) (plumbing.EncodedObject, error) {
	obj, ok := tx.Objects[h]
	if !ok || (plumbing.AnyObject != t && obj.Type() != t) {
		return nil, plumbing.ErrObjectNotFound
	}

	return obj, nil
}

func (tx *TxObjectStorage) Commit() error {
	for h, obj := range tx.Objects {
		delete(tx.Objects, h)
		if _, err := tx.Storage.SetEncodedObject(obj); err != nil {
			return err
		}
	}

	return nil
}

func (tx *TxObjectStorage) Rollback() error {
	tx.Objects = make(map[common.ObjectHash]plumbing.EncodedObject)
	return nil
}

type ReferenceStorage map[plumbing.ReferenceName]*plumbing.Reference

func (r ReferenceStorage) SetReference(ref *plumbing.Reference) error {
	if ref != nil {
		r[ref.Name()] = ref
	}

	return nil
}

func (r ReferenceStorage) CheckAndSetReference(ref, old *plumbing.Reference) error {
	if ref == nil {
		return nil
	}

	if old != nil {
		tmp := r[ref.Name()]
		if tmp != nil && tmp.Hash() != old.Hash() {
			return storage.ErrReferenceHasChanged
		}
	}
	r[ref.Name()] = ref
	return nil
}

func (r ReferenceStorage) Reference(n plumbing.ReferenceName) (*plumbing.Reference, error) {
	ref, ok := r[n]
	if !ok {
		return nil, plumbing.ErrReferenceNotFound
	}

	return ref, nil
}

func (r ReferenceStorage) IterReferences() (storer.ReferenceIter, error) {
	var refs []*plumbing.Reference
	for _, ref := range r {
		refs = append(refs, ref)
	}

	return storer.NewReferenceSliceIter(refs), nil
}

func (r ReferenceStorage) CountLooseRefs() (int, error) {
	return len(r), nil
}

func (r ReferenceStorage) PackRefs() error {
	return nil
}

func (r ReferenceStorage) RemoveReference(n plumbing.ReferenceName) error {
	delete(r, n)
	return nil
}

type ShallowStorage []common.ObjectHash

func (s *ShallowStorage) SetShallow(commits []common.ObjectHash) error {
	*s = commits
	return nil
}

func (s ShallowStorage) Shallow() ([]common.ObjectHash, error) {
	return s, nil
}

type ModuleStorage map[string]*Storage

func (s ModuleStorage) Module(name string) (storage.Storer, error) {
	if m, ok := s[name]; ok {
		return m, nil
	}

	m := NewStorage()
	s[name] = m

	return m, nil
}
