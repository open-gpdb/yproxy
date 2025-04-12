package backups

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type StorageBackupInteractor struct {
	Storage storage.StorageInteractor
}

// get lsn of the oldest backup
func (b *StorageBackupInteractor) GetFirstLSN(seg uint64) (uint64, error) {
	objects, err := b.Storage.ListPath(fmt.Sprintf("segments_005/seg%d/basebackups_005/", seg), true, nil)
	if err != nil {	
		ylogger.Zero.Debug().Err(err).Msg("GetFirstLSN: list result")
		return 0, err
	}
	ylogger.Zero.Debug().Int("size", len(objects)).Msg("GetFirstLSN: list result size")

	minLSN := BackupLSN{Lsn: ^uint64(0)}
	for _, obj := range objects {
		ylogger.Zero.Debug().Str("path", obj.Path).Msg("GetFirstLSN: checking")
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

		ylogger.Zero.Debug().Str("path", obj.Path).Uint64("lsn", lsn.Lsn).Msg("GetFirstLSN: parsed backup lsn")

		if lsn.Lsn < minLSN.Lsn {
			minLSN.Lsn = lsn.Lsn
		}
	}

	return minLSN.Lsn, err
}
