package dotgit

import (
	"fmt"
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
)

type packWriterSequential struct {
	Notify func(plumbing.Hash, *idxfile.Writer)

	fs        billy.Filesystem
	tmp       billy.File
	checksum  plumbing.Hash
	parser    *packfile.Parser
	idxWriter *idxfile.Writer
	result    error
}

func newPackWriterSequential(fs billy.Filesystem) (*packWriterSequential, error) {
	tempPack, err := fs.TempFile(fs.Join(objectsPath, packPath), "tmp_pack_")
	if err != nil {
		return nil, err
	}

	writer := &packWriterSequential{
		fs:  fs,
		tmp: tempPack,
	}

	return writer, nil
}

func (w *packWriterSequential) Write(p []byte) (int, error) {
	return w.tmp.Write(p)
}

// Close closes all the file descriptors and save the final packfile, if nothing
// was written, the tempfiles are deleted without writing a packfile.
func (w *packWriterSequential) Close() error {
	defer func() {
		if w.Notify != nil && w.idxWriter != nil && w.idxWriter.Finished() {
			w.Notify(w.checksum, w.idxWriter)
		}
	}()

	err := w.tmp.Close()
	if err != nil {
		return err
	}

	fw, err := w.fs.Open(w.tmp.Name())
	if err != nil {
		return err
	}

	w.idxWriter = new(idxfile.Writer)
	parser := packfile.NewParser(fw, packfile.WithScannerObservers(w.idxWriter))
	h, err := parser.Parse()
	if err != nil {
		return err
	}
	w.checksum = h

	if !w.idxWriter.Finished() {
		return w.clean()
	}

	return w.save()
}

func (w *packWriterSequential) clean() error {
	return w.fs.Remove(w.tmp.Name())
}

func (w *packWriterSequential) save() error {
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

	return w.fs.Rename(w.tmp.Name(), fmt.Sprintf("%s.pack", base))
}

func (w *packWriterSequential) encodeIdx(writer io.Writer) error {
	idx, err := w.idxWriter.Index()
	if err != nil {
		return err
	}

	e := idxfile.NewEncoder(writer)
	_, err = e.Encode(idx)
	return err
}
