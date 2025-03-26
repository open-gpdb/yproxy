package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/settings"
)

// Storage prefix uses as path to folder.
// WARNING "/path/to/folder" dont work
// "/path/to/folder/" + "path/to/file.txt"
type FileStorageInteractor struct {
	StorageInteractor
	cnf *config.Storage
}

func (s *FileStorageInteractor) CatFileFromStorage(name string, offset int64, _ []settings.StorageSettings) (io.ReadCloser, error) {
	file, err := os.Open(path.Join(s.cnf.StoragePrefix, name))
	if err != nil {
		return nil, err
	}
	_, err = io.CopyN(io.Discard, file, offset)
	return file, err
}
func (s *FileStorageInteractor) ListPath(prefix string, _ bool, _ []settings.StorageSettings) ([]*object.ObjectInfo, error) {
	var data []*object.ObjectInfo
	err := filepath.WalkDir(s.cnf.StoragePrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		fileinfo, err := file.Stat()
		if err != nil {
			return err
		}

		cuttedPrefix, _ := strings.CutPrefix(prefix, "/")
		neededPrefix := s.cnf.StoragePrefix + cuttedPrefix
		if !strings.HasPrefix(path, neededPrefix) {
			return nil
		}
		cPath, ok := strings.CutPrefix(path, s.cnf.StoragePrefix)

		if !ok {
			return err
		}
		data = append(data, &object.ObjectInfo{Path: "/" + cPath, Size: fileinfo.Size()})
		return nil
	})
	return data, err
}

func (s *FileStorageInteractor) PutFileToDest(name string, r io.Reader, _ []settings.StorageSettings) error {
	fPath := path.Join(s.cnf.StoragePrefix, name)
	fDir := path.Dir(fPath)
	os.MkdirAll(fDir, 0700)
	file, err := os.Create(fPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r)
	return err
}

func (s *FileStorageInteractor) PatchFile(name string, r io.ReadSeeker, startOffset int64) error {
	//UNUSED TODO
	return fmt.Errorf("TODO")
}

func (s *FileStorageInteractor) MoveObject(from string, to string) error {
	fromPath := path.Join(s.cnf.StoragePrefix, from)
	toPath := path.Join(s.cnf.StoragePrefix, to)
	toDir := path.Dir(toPath)
	os.MkdirAll(toDir, 0700)
	return os.Rename(fromPath, toPath)
}

func (s *FileStorageInteractor) CopyObject(from, to string) error {
	fromPath := path.Join(s.cnf.StoragePrefix, from)
	toPath := path.Join(s.cnf.StoragePrefix, to)
	toDir := path.Dir(toPath)
	os.MkdirAll(toDir, 0700)
	fromFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	toFile, err := os.OpenFile(toPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = io.Copy(fromFile, toFile)
	return err
}

func (s *FileStorageInteractor) DeleteObject(key string) error {
	return os.Remove(path.Join(s.cnf.StoragePrefix, key))
}

func (s *FileStorageInteractor) AbortMultipartUploads() error {
	return nil
}

func (s *FileStorageInteractor) AbortMultipartUpload(key, uploadId string) error {
	return nil
}

func (s *FileStorageInteractor) ListFailedMultipartUploads() (map[string]string, error) {
	return nil, nil
}
