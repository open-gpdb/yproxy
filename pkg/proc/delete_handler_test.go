package proc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/message"
	mock "github.com/yezzey-gp/yproxy/pkg/mock"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/proc"
	"go.uber.org/mock/gomock"
)

func TestFilesToDeletion(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.DeleteMessage{
		Name:    "path",
		Port:    6000,
		Segnum:  0,
		Confirm: false,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "1663_16530_not-deleted_18002_"},
		{Path: "1663_16530_deleted-after-backup_18002_"},
		{Path: "1663_16530_deleted-when-backup-start_18002_"},
		{Path: "1663_16530_deleted-before-backup_18002_"},
		{Path: "some_trash"},
	}
	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("", msg.Name, true).Return(filesInStorage, nil)

	backup := mock.NewMockBackupInterractor(ctrl)
	backup.EXPECT().GetFirstLSN(msg.Segnum).Return(uint64(1337), nil)

	vi := map[string]bool{
		"1663_16530_not-deleted_18002_":               true,
		"1663_16530_deleted-after-backup_18002_":      true,
		"1663_16530_deleted-when-backup-start_18002_": true,
	}
	ei := map[string]uint64{
		"1663_16530_deleted-after-backup_18002_":      uint64(1400),
		"1663_16530_deleted-when-backup-start_18002_": uint64(1337),
		"1663_16530_deleted-before-backup_18002_":     uint64(1300),
	}
	database := mock.NewMockDatabaseInterractor(ctrl)
	database.EXPECT().GetVirtualExpireIndexes(msg.Port).Return(vi, ei, nil)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		DbInterractor:      database,
		BackupInterractor:  backup,
		Cnf:                &config.Vacuum{CheckBackup: true},
	}

	list, err := handler.ListGarbageFiles("", msg)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, "1663_16530_deleted-before-backup_18002_", list[0])
	assert.Equal(t, "some_trash", list[1])
}

func TestTrashPathConversion(t *testing.T) {

	type tt struct {
		Path         string
		ExpTrashPath string
	}

	for _, tc := range []tt{
		{
			Path:         "segments_005/seg60/basebackups_005/yezzey/1663_16712_001128848b0d46158f270005bc9cd82a_28144635_2305__DY_1_xlog_109463918313472",
			ExpTrashPath: "trash/segments_005/seg60/basebackups_005/yezzey/1663_16712_001128848b0d46158f270005bc9cd82a_28144635_2305__DY_1_xlog_109463918313472",
		},
	} {

		resP := proc.TrashPathFromRegPath(tc.Path, 60)

		assert.Equal(t, tc.ExpTrashPath, resP)

		resP2 := proc.RegPathFromTrasnPath(resP, 60)

		assert.Equal(t, tc.Path, resP2)
	}
}
