package dotgit

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/hash"

	"github.com/go-git/go-billy/v5"
)

// PackWriter is a io.Writer that generates the packfile index simultaneously,
// a packfile.Decoder is used with a file reader to read the file being written
// this operation is synchronized with the write operations.
// The packfile is written in a temp file, when Close is called this file
// is renamed/moved (depends on the Filesystem implementation) to the final
// location, if the PackWriter is not used, nothing is written.
type PackWriter struct {
	Notify func(plumbing.Hash, *idxfile.Writer)

	fs   billy.Filesystem
	temp billy.File
	pr   *io.PipeReader
	pw   *io.PipeWriter

	checksum  plumbing.Hash
	parser    *packfile.Parser
	idxWriter *idxfile.Writer
	result    error

	mw io.Writer
	wg sync.WaitGroup
}

func newPackWrite(fs billy.Filesystem) (*PackWriter, error) {
	tempPack, err := fs.TempFile(fs.Join(objectsPath, packPath), "tmp_pack_")
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	writer := &PackWriter{
		fs:   fs,
		temp: tempPack,

		idxWriter: new(idxfile.Writer),
		mw:        io.MultiWriter(pw, tempPack),
		pr:        pr,
		pw:        pw,

		wg: sync.WaitGroup{},
	}
	writer.wg.Add(1)

	go func() {
		defer pw.Close()
		defer writer.wg.Done()

		writer.parser = packfile.NewParser(pr,
			packfile.WithScannerObservers(writer.idxWriter),
		)
		h, err := writer.parser.Parse()
		if err != nil {
			if !errors.Is(err, packfile.ErrEmptyPackfile) {
				writer.result = err
			}
			return
		}

		writer.checksum = h
	}()

	return writer, nil
}

func (w *PackWriter) Write(p []byte) (int, error) {
	return w.mw.Write(p)
}

// Close closes all the file descriptors and save the final packfile, if nothing
// was written, the tempfiles are deleted without writing a packfile.
func (w *PackWriter) Close() error {
	defer func() {
		if w.Notify != nil && w.idxWriter != nil && w.idxWriter.Finished() {
			w.Notify(w.checksum, w.idxWriter)
		}
	}()

	w.wg.Wait()
	err := w.pr.Close()
	if err != nil {
		return err
	}
	if !w.idxWriter.Finished() {
		return w.clean()
	}

	return w.save()
}

func (w *PackWriter) clean() error {
	return w.fs.Remove(w.temp.Name())
}

func (w *PackWriter) save() error {
	base := w.fs.Join(objectsPath, packPath, fmt.Sprintf("pack-%s", w.checksum))
	idx, err := w.fs.Create(fmt.Sprintf("%s.idx", base))
	if err != nil {
		return err
	}

	if err := w.encodeIdx(idx); err != nil {
		return err
	}

	if err := idx.Close(); err != nil {
		return err
	}

	return w.fs.Rename(w.temp.Name(), fmt.Sprintf("%s.pack", base))
}

func (w *PackWriter) encodeIdx(writer io.Writer) error {
	idx, err := w.idxWriter.Index()
	if err != nil {
		return err
	}

	e := idxfile.NewEncoder(writer)
	_, err = e.Encode(idx)
	return err
}

type ObjectWriter struct {
	objfile.Writer
	fs billy.Filesystem
	f  billy.File
}

func newObjectWriter(fs billy.Filesystem) (*ObjectWriter, error) {
	f, err := fs.TempFile(fs.Join(objectsPath, packPath), "tmp_obj_")
	if err != nil {
		return nil, err
	}

	return &ObjectWriter{
		Writer: (*objfile.NewWriter(f)),
		fs:     fs,
		f:      f,
	}, nil
}

func (w *ObjectWriter) Close() error {
	if err := w.Writer.Close(); err != nil {
		return err
	}

	if err := w.f.Close(); err != nil {
		return err
	}

	return w.save()
}

func (w *ObjectWriter) save() error {
	hex := w.Hash().String()
	file := w.fs.Join(objectsPath, hex[0:2], hex[2:hash.HexSize])

	return w.fs.Rename(w.f.Name(), file)
}
