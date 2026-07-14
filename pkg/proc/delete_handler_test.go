package proc_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
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
	assert.Equal(t, "1663_16530_deleted-before-backup_18002_", list[0].Path)
	assert.Equal(t, "some_trash", list[1].Path)
}

func TestFilesToDeletionSkipsRecentlyCreatedFiles(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.DeleteMessage{
		Name:    "path",
		Port:    6000,
		Segnum:  0,
		Confirm: false,
	}

	oldFile := &object.ObjectInfo{
		Path:    "1663_16530_old-garbage_18002_",
		LastMod: time.Now().Add(-time.Hour),
	}
	recentFile := &object.ObjectInfo{
		Path:    "1663_16530_recently-created_18002_",
		LastMod: time.Now().Add(time.Hour), // created after vacuum started
	}

	filesInStorage := []*object.ObjectInfo{oldFile, recentFile}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("", msg.Name, true).Return(filesInStorage, nil)

	backup := mock.NewMockBackupInterractor(ctrl)
	backup.EXPECT().GetFirstLSN(msg.Segnum).Return(uint64(1337), nil)

	// Neither file is present in virtual or expire index, so both would
	// normally be considered garbage.
	vi := map[string]bool{}
	ei := map[string]uint64{}
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
	assert.Equal(t, 1, len(list))
	assert.Equal(t, oldFile.Path, list[0].Path)
}

func TestFilesToDeletionRespectsProtectionSecondsWindow(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.DeleteMessage{
		Name:    "path",
		Port:    6000,
		Segnum:  0,
		Confirm: false,
	}

	oldFile := &object.ObjectInfo{
		Path:    "1663_16530_old-garbage_18002_",
		LastMod: time.Now().Add(-2 * time.Hour),
	}
	// Created before the vacuum procedure started, but within the
	// configured protection window, so it must be skipped.
	withinWindowFile := &object.ObjectInfo{
		Path:    "1663_16530_within-window_18002_",
		LastMod: time.Now().Add(-10 * time.Minute),
	}

	filesInStorage := []*object.ObjectInfo{oldFile, withinWindowFile}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("", msg.Name, true).Return(filesInStorage, nil)

	backup := mock.NewMockBackupInterractor(ctrl)
	backup.EXPECT().GetFirstLSN(msg.Segnum).Return(uint64(1337), nil)

	// Neither file is present in virtual or expire index, so both would
	// normally be considered garbage.
	vi := map[string]bool{}
	ei := map[string]uint64{}
	database := mock.NewMockDatabaseInterractor(ctrl)
	database.EXPECT().GetVirtualExpireIndexes(msg.Port).Return(vi, ei, nil)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		DbInterractor:      database,
		BackupInterractor:  backup,
		Cnf: &config.Vacuum{
			CheckBackup:      true,
			ProtectionWindow: time.Hour,
		},
	}

	list, err := handler.ListGarbageFiles("", msg)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, oldFile.Path, list[0].Path)
}

func TestFilesToDeletionClampsNegativeProtectionSeconds(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.DeleteMessage{
		Name:    "path",
		Port:    6000,
		Segnum:  0,
		Confirm: false,
	}

	oldFile := &object.ObjectInfo{
		Path:    "1663_16530_old-garbage_18002_",
		LastMod: time.Now().Add(-2 * time.Hour),
	}
	// Created after the vacuum procedure started. Even with a negative
	// (misconfigured) ProtectionSeconds, this file must still be
	// unconditionally protected.
	recentFile := &object.ObjectInfo{
		Path:    "1663_16530_recently-created_18002_",
		LastMod: time.Now().Add(time.Hour),
	}

	filesInStorage := []*object.ObjectInfo{oldFile, recentFile}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("", msg.Name, true).Return(filesInStorage, nil)

	backup := mock.NewMockBackupInterractor(ctrl)
	backup.EXPECT().GetFirstLSN(msg.Segnum).Return(uint64(1337), nil)

	vi := map[string]bool{}
	ei := map[string]uint64{}
	database := mock.NewMockDatabaseInterractor(ctrl)
	database.EXPECT().GetVirtualExpireIndexes(msg.Port).Return(vi, ei, nil)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		DbInterractor:      database,
		BackupInterractor:  backup,
		Cnf: &config.Vacuum{
			CheckBackup:      true,
			ProtectionWindow: -time.Hour, // misconfigured, must be clamped to 0
		},
	}

	list, err := handler.ListGarbageFiles("", msg)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, oldFile.Path, list[0].Path)
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

func TestListDelete2Files(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "1663_16530_not-deleted_18002_"},
		{Path: "1663_16530_deleted-after-backup_18002_"},
		{Path: "1663_16530_deleted-when-backup-start_18002_"},
		{Path: "1663_16530_deleted-before-backup_18002_"},
		{Path: "some_trash"},
	}
	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("", msg.Prefix, true).Return(filesInStorage, nil)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		DbInterractor:      nil,
		BackupInterractor:  nil,
		Cnf:                &config.Vacuum{CheckBackup: true},
	}

	actualFilesToDelete, err := handler.ListDelete2Files("", msg)
	assert.NoError(t, err)
	assert.Equal(t, len(filesInStorage), len(actualFilesToDelete))
	assert.Equal(t, filesInStorage, actualFilesToDelete)
}

func TestDeletePrefixInBucketDeletesAllFilesOnceInParallelGarbagePass(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "trash/a"},
		{Path: "trash/b"},
		{Path: "trash/c"},
		{Path: "trash/d"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("bucket", msg.Prefix, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("bucket").Return(map[string]string{}, nil)

	var mu sync.Mutex
	deleted := make(map[string]int)
	storage.EXPECT().DeleteObject("bucket", gomock.Any()).DoAndReturn(func(bucket, key string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[key]++
		return nil
	}).Times(len(filesInStorage))

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		Cnf:                &config.Vacuum{TrashDeleteWorkers: 4},
	}

	err := handler.DeletePrefixInBucket("bucket", msg)
	assert.NoError(t, err)
	for _, file := range filesInStorage {
		assert.Equal(t, 1, deleted[file.Path], fmt.Sprintf("unexpected delete count for %s", file.Path))
	}
}

func TestDeletePrefixInBucketRetriesFailedGarbageDeletes(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "trash/a"},
		{Path: "trash/b"},
		{Path: "trash/c"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("bucket", msg.Prefix, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("bucket").Return(map[string]string{}, nil)

	var mu sync.Mutex
	attempts := make(map[string]int)
	storage.EXPECT().DeleteObject("bucket", gomock.Any()).DoAndReturn(func(bucket, key string) error {
		mu.Lock()
		defer mu.Unlock()
		attempts[key]++
		if key == "trash/b" && attempts[key] == 1 {
			return errors.New("transient delete failure")
		}
		return nil
	}).Times(4)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		Cnf:                &config.Vacuum{TrashDeleteWorkers: 3},
	}

	err := handler.DeletePrefixInBucket("bucket", msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts["trash/a"])
	assert.Equal(t, 2, attempts["trash/b"])
	assert.Equal(t, 1, attempts["trash/c"])
}

func TestDeletePrefixInBucketReturnsFailedGarbageDeletesAfterRetries(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "trash/a"},
		{Path: "trash/b"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("bucket", msg.Prefix, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("bucket").Return(map[string]string{}, nil)
	storage.EXPECT().DeleteObject("bucket", gomock.Any()).DoAndReturn(func(bucket, key string) error {
		if key == "trash/b" {
			return errors.New("persistent delete failure")
		}
		return nil
	}).Times(11)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		Cnf:                &config.Vacuum{TrashDeleteWorkers: 2},
	}

	err := handler.DeletePrefixInBucket("bucket", msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete some files")
}

func TestDeletePrefixInBucketCapsWorkerCountToFileCount(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "trash/a"},
		{Path: "trash/b"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("bucket", msg.Prefix, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("bucket").Return(map[string]string{}, nil)

	var mu sync.Mutex
	deleted := make(map[string]int)
	storage.EXPECT().DeleteObject("bucket", gomock.Any()).DoAndReturn(func(bucket, key string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[key]++
		return nil
	}).Times(len(filesInStorage))

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		Cnf:                &config.Vacuum{TrashDeleteWorkers: 10},
	}

	err := handler.DeletePrefixInBucket("bucket", msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted["trash/a"])
	assert.Equal(t, 1, deleted["trash/b"])
}

func TestDeletePrefixInBucketUsesDefaultWorkerCountWhenConfiguredZero(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.Delete2Message{
		Prefix:  "trash",
		Garbage: true,
		Confirm: true,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "trash/a"},
		{Path: "trash/b"},
		{Path: "trash/c"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("bucket", msg.Prefix, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("bucket").Return(map[string]string{}, nil)

	var mu sync.Mutex
	deleted := make(map[string]int)
	storage.EXPECT().DeleteObject("bucket", gomock.Any()).DoAndReturn(func(bucket, key string) error {
		mu.Lock()
		defer mu.Unlock()
		deleted[key]++
		return nil
	}).Times(len(filesInStorage))

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		Cnf:                &config.Vacuum{TrashDeleteWorkers: 0},
	}

	err := handler.DeletePrefixInBucket("bucket", msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted["trash/a"])
	assert.Equal(t, 1, deleted["trash/b"])
	assert.Equal(t, 1, deleted["trash/c"])
}

func TestDeleteGarbageInBucketMovesObjectsWhenCrazyDropDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)

	msg := message.DeleteMessage{
		Name:      "path",
		Port:      6000,
		Segnum:    0,
		Confirm:   true,
		CrazyDrop: false,
	}

	filesInStorage := []*object.ObjectInfo{
		{Path: "segments_005/seg0/basebackups_005/yezzey/file1"},
		{Path: "segments_005/seg0/basebackups_005/yezzey/file2"},
	}

	storage := mock.NewMockStorageInteractor(ctrl)
	storage.EXPECT().ListBucketPath("trash", msg.Name, true).Return(filesInStorage, nil)
	storage.EXPECT().ListFailedMultipartUploads("trash").Return(map[string]string{}, nil)
	storage.EXPECT().MoveObject("trash", filesInStorage[0].Path, proc.TrashPathFromRegPath(filesInStorage[0].Path, int(msg.Segnum))).Return(nil)
	storage.EXPECT().MoveObject("trash", filesInStorage[1].Path, proc.TrashPathFromRegPath(filesInStorage[1].Path, int(msg.Segnum))).Return(nil)

	database := mock.NewMockDatabaseInterractor(ctrl)
	database.EXPECT().GetVirtualExpireIndexes(msg.Port).Return(map[string]bool{}, map[string]uint64{
		filesInStorage[0].Path: 0,
		filesInStorage[1].Path: 0,
	}, nil)

	handler := proc.BasicGarbageMgr{
		StorageInterractor: storage,
		DbInterractor:      database,
		BackupInterractor:  nil,
		Cnf:                &config.InstanceConfig().VacuumCnf,
	}

	cnfBackup := *config.InstanceConfig()
	defer func() {
		*config.InstanceConfig() = cnfBackup
	}()
	config.EmbedDefaults(config.InstanceConfig())

	err := handler.DeleteGarbageInBucket("trash", msg)
	assert.NoError(t, err)
}
