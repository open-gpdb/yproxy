package backups

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yezzey-gp/yproxy/pkg/storage"
)

type S3BackupInteractor struct {
	Storage storage.StorageInteractor
}

// get lsn of the oldest backup
func (b *S3BackupInteractor) GetFirstLSN(seg uint64) (uint64, error) {
	objects, err := b.Storage.ListPath(fmt.Sprintf("segments_005/seg%d/basebackups_005/", seg), true)
	if err != nil {
		return 0, err
	}

	minLSN := BackupLSN{Lsn: ^uint64(0)}
	for _, obj := range objects {
		if !strings.Contains(obj.Path, ".json") {
			continue
		}

		// cat file
		reader, err := b.Storage.CatFileFromStorage(obj.Path, 0, nil)
		if err != nil {
			continue
		}
		content, err := io.ReadAll(reader)
		if err != nil {
			continue
		}

		lsn := BackupLSN{}
		err = json.Unmarshal(content, &lsn)
		if err != nil {
			continue
		}

		if lsn.Lsn < minLSN.Lsn {
			minLSN.Lsn = lsn.Lsn
		}
	}

	return minLSN.Lsn, err
}
