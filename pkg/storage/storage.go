package storage

import (
	"fmt"
	"io"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/tablespace"
)

type StorageReader interface {
	CatFileFromStorage(name string, offset int64, setts []settings.StorageSettings) (io.ReadCloser, error)
}

type StorageWriter interface {
	PutFileToDest(name string, r io.Reader, settings []settings.StorageSettings) error
	PatchFile(name string, r io.ReadSeeker, startOffset int64) error
}

type StorageLister interface {
	ListPath(prefix string, useCache bool, settings []settings.StorageSettings) ([]*object.ObjectInfo, error)
	ListFailedMultipartUploads() (map[string]string, error)
}

type StorageMover interface {
	MoveObject(from string, to string) error
	DeleteObject(key string) error
	AbortMultipartUpload(key, uploadId string) error
}

type StorageCopier interface {
	CopyObject(from, to, fromStoragePrefix, fromStorageBucket string) error
}

//go:generate mockgen -destination=pkg/mock/storage.go -package=mock
type StorageInteractor interface {
	StorageReader
	StorageWriter
	StorageLister
	StorageMover
	StorageCopier
}

func NewStorage(cnf *config.Storage) (StorageInteractor, error) {
	switch cnf.StorageType {
	case "fs":
		return &FileStorageInteractor{
			cnf: cnf,
		}, nil
	case "s3":
		return &S3StorageInteractor{
			pool:          NewSessionPool(cnf),
			cnf:           cnf,
			bucketMap:     buildBucketMapFromCnf(cnf),
			credentialMap: buildCredMapFromCnf(cnf),
		}, nil
	default:
		return nil, fmt.Errorf("wrong storage type %s", cnf.StorageType)
	}
}

func buildBucketMapFromCnf(cnf *config.Storage) map[string]string {
	mp := cnf.TablespaceMap
	if mp == nil {
		/* fallback for backward-compatibility if to TableSpace map configured */
		mp = map[string]string{}
	}
	if _, ok := mp[tablespace.DefaultTableSpace]; !ok {
		mp[tablespace.DefaultTableSpace] = cnf.StorageBucket
	}
	return mp
}

func buildCredMapFromCnf(cnf *config.Storage) map[string]config.StorageCredentials {
	mp := cnf.CredentialMap
	if mp == nil {
		/* fallback for backward-compatibility if to TableSpace map configured */
		mp = map[string]config.StorageCredentials{}
	}
	if _, ok := mp[cnf.StorageBucket]; !ok {
		mp[cnf.StorageBucket] = config.StorageCredentials{
			AccessKeyId:     cnf.AccessKeyId,
			SecretAccessKey: cnf.SecretAccessKey,
		}
	}
	return mp
}
