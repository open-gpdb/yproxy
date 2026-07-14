package proc

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/backups"
	"github.com/yezzey-gp/yproxy/pkg/database"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
	"golang.org/x/time/rate"
)

//go:generate mockgen -destination=../../../test/mocks/mock_object.go -package mocks -build_flags -mod=readonly github.com/wal-g/wal-g/pkg/storages/storage Object
type GarbageMgr interface {
	HandleDeleteGarbage(message.DeleteMessage) error
	HandleDeleteFile(message.DeleteMessage) error
	HandleUntrashifyFile(message.UntrashifyMessage) error
}

type BasicGarbageMgr struct {
	BackupInterractor  backups.BackupInterractor
	DbInterractor      database.DatabaseInterractor
	StorageInterractor storage.StorageInteractor

	Cnf *config.Vacuum
}

var _ GarbageMgr = &BasicGarbageMgr{}

func TrashPathFromRegPath(p string, segnum int) string {

	filePathParts := strings.Split(p, "/")

	destPath := path.Join(
		"trash",
		"segments_005",
		fmt.Sprintf("seg%d", segnum),
		"basebackups_005",
		"yezzey", filePathParts[len(filePathParts)-1])

	return destPath
}

func RegPathFromTrasnPath(p string, segnum int) string {

	filePathParts := strings.Split(p, "/")

	destPath := path.Join(
		"segments_005",
		fmt.Sprintf("seg%d", segnum),
		"basebackups_005",
		"yezzey", filePathParts[len(filePathParts)-1])

	return destPath
}

// HandleUntrashifyFile implements GarbageMgr.
func (dh *BasicGarbageMgr) HandleUntrashifyFile(msg message.UntrashifyMessage) error {

	ylogger.Zero.Info().Str("path", msg.Name).Msg("listing prefix")
	objectMetas, err := dh.StorageInterractor.ListPath(msg.Name, true, nil)
	if err != nil {
		return errors.Wrap(err, "could not list objects")
	}

	for _, file := range objectMetas {
		ylogger.Zero.Info().Str("file", file.Path).Str("dest-path", RegPathFromTrasnPath(file.Path, int(msg.Segnum))).Msg("file will be untrashified")
	}

	if !msg.Confirm { //do not delete files if no confirmation flag provided
		return nil
	}

	for _, file := range objectMetas {
		tp := RegPathFromTrasnPath(file.Path, int(msg.Segnum))
		/* XXX: fix this */
		err = dh.StorageInterractor.MoveObject(dh.StorageInterractor.DefaultBucket(), file.Path, tp)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
 * The design looks an awkward,
 * because the function depends on two different vacuum config:
 * 	- the local dh.Cnf
 * 	- the global config.InstanceConfig()
 * Example: TestDeleteGarbageInBucketMovesObjectsWhenCrazyDropDisabled
 */
func (dh *BasicGarbageMgr) DeleteGarbageInBucket(bucket string, msg message.DeleteMessage) error {
	fileList, err := dh.ListGarbageFiles(bucket, msg)
	if err != nil {
		return errors.Wrap(err, "failed to delete file")
	}
	uploads, err := dh.StorageInterractor.ListFailedMultipartUploads(bucket)
	if err != nil {
		return err
	}
	ylogger.Zero.Info().Str("bucket", bucket).Int("amount", len(uploads)).Msg("multipart uploads will be aborted")

	for _, file := range fileList {
		ylogger.Zero.Info().Str("bucket", bucket).Bool("crazy mode", msg.CrazyDrop).Str("file", file.Path).Msg("file will be deleted")
	}
	for _, upload := range uploads {
		ylogger.Zero.Info().Str("bucket", bucket).Str("uploadId", upload).Msg("upload will be aborted")
	}

	if !msg.Confirm { // Do not delete files if no confirmation flag provided
		ylogger.Zero.Info().Msg("do not perform actual delete files as no confirmation flag provided")
		return nil
	}

	/*
	 * Burst at 20% of vacuum rate capacity. It is pretty arbitrary at this time,
	 * but its not like something we need config field for...
	 */
	limRate := config.InstanceConfig().VacuumCnf.FileChunkPerSec
	limiter := rate.NewLimiter(rate.Limit(limRate), limRate/5)
	ctx := context.Background()

	var failedActionMsg, failedFilesMsg string
	var operate func(file *object.ObjectInfo) error
	if msg.CrazyDrop {
		failedActionMsg = "failed to delete some files"
		failedFilesMsg = "some files were not deleted"
		operate = func(file *object.ObjectInfo) error {
			ylogger.Zero.Info().Str("bucket", bucket).Str("path", file.Path).Msg("immediately delete garbage file")
			return dh.StorageInterractor.DeleteObject(bucket, file.Path)
		}
	} else {
		failedActionMsg = "failed to move some files"
		failedFilesMsg = "some files were not moved"
		operate = func(file *object.ObjectInfo) error {
			tp := TrashPathFromRegPath(file.Path, int(msg.Segnum))
			ylogger.Zero.Info().Str("bucket", bucket).Str("path", file.Path).Msg("move garbage file to trash")
			return dh.StorageInterractor.MoveObject(bucket, file.Path, tp)
		}
	}

	var failed []*object.ObjectInfo
	for retryCount := 0; len(fileList) > 0 && retryCount < 10; retryCount++ {
		for _, file := range fileList {
			/* Don't move too fast */
			if err := limiter.Wait(ctx); err != nil {
				break
			}
			err = operate(file)
			if err != nil {
				ylogger.Zero.Warn().AnErr("err", err).Str("bucket", bucket).Str("file", file.Path).Msg(failedActionMsg)
				failed = append(failed, file)
			}
		}
		fileList = failed
		failed = make([]*object.ObjectInfo, 0)
	}

	if len(fileList) > 0 {
		ylogger.Zero.Error().Str("bucket", bucket).Int("failed files count", len(fileList)).Msg(failedFilesMsg)
		ylogger.Zero.Error().Str("bucket", bucket).Any("failed files", fileList).Msg(failedActionMsg)
		return errors.Wrap(err, failedActionMsg)
	}

	for key, uploadId := range uploads {
		/* Don't move too fast */
		if err := limiter.Wait(ctx); err != nil {
			break
		}
		if err := dh.StorageInterractor.AbortMultipartUpload(bucket, key, uploadId); err != nil {
			return err
		}
	}

	return nil
}

func (dh *BasicGarbageMgr) HandleDeleteGarbage(msg message.DeleteMessage) error {
	for _, b := range dh.StorageInterractor.ListBuckets() {
		if err := dh.DeleteGarbageInBucket(b, msg); err != nil {
			return err
		}
	}
	return nil
}
func (dh *BasicGarbageMgr) ListDelete2Files(bucket string, msg message.Delete2Message) ([]*object.ObjectInfo, error) {
	// Get first backup lsn
	var err error

	// List files in storage
	ylogger.Zero.Info().Str("path", msg.Prefix).Msg("listing prefix")
	objectMetas, err := dh.StorageInterractor.ListBucketPath(bucket, msg.Prefix, true)
	if err != nil {
		return nil, errors.Wrap(err, "could not list objects")
	}
	ylogger.Zero.Info().Int("amount", len(objectMetas)).Msg("objects count")

	filesToDelete := make([]*object.ObjectInfo, 0)
	for i := range objectMetas {
		ylogger.Zero.Debug().Str("file", objectMetas[i].Path).
			Str("will be deleted", msg.Prefix)

		filesToDelete = append(filesToDelete, objectMetas[i])

	}

	ylogger.Zero.Info().Int("amount", len(filesToDelete)).Msg("files will be deleted")

	return filesToDelete, nil
}

func (dh *BasicGarbageMgr) garbageTrashParallel(bucket string, fileList []*object.ObjectInfo) ([]*object.ObjectInfo, error) {
	workerCount := dh.Cnf.TrashDeleteWorkers
	if workerCount <= 0 {
		workerCount = config.DefaultTrashDeleteWorkers
	}
	if workerCount > len(fileList) {
		workerCount = len(fileList)
	}
	if workerCount == 0 {
		return nil, nil
	}

	jobs := make(chan *object.ObjectInfo)
	failedCh := make(chan *object.ObjectInfo, len(fileList))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Go(func() {
			for file := range jobs {
				if err := dh.StorageInterractor.DeleteObject(bucket, file.Path); err != nil {
					ylogger.Zero.Warn().AnErr("err", err).Str("bucket", bucket).Str("file", file.Path).Msg("failed to delete garbage file")
					failedCh <- file
				}
			}
		})
	}

	for _, file := range fileList {
		jobs <- file
	}
	close(jobs)

	wg.Wait()
	close(failedCh)

	failed := make([]*object.ObjectInfo, 0, len(fileList))
	for file := range failedCh {
		failed = append(failed, file)
	}
	if len(failed) > 0 {
		return failed, errors.New("failed to delete some garbage files")
	}

	return nil, nil
}

func (dh *BasicGarbageMgr) DeletePrefixInBucket(bucket string, msg message.Delete2Message) error {
	fileList, err := dh.ListDelete2Files(bucket, msg) // Return the list of files to be deleted
	if err != nil {
		return errors.Wrap(err, "failed to delete file")
	}
	uploads, err := dh.StorageInterractor.ListFailedMultipartUploads(bucket)
	if err != nil {
		return err
	}
	ylogger.Zero.Info().Str("bucket", bucket).Int("amount", len(uploads)).Msg("multipart uploads will be aborted")

	for _, file := range fileList {
		ylogger.Zero.Info().Str("bucket", bucket).Str("file", file.Path).Msg("file will be deleted")
	}
	for _, upload := range uploads {
		ylogger.Zero.Info().Str("bucket", bucket).Str("uploadId", upload).Msg("upload will be aborted")
	}

	if !msg.Confirm { // Do not delete files if no confirmation flag provided
		ylogger.Zero.Info().Msg("do not perform actual delete files as no confirmation flag provided")
		return nil
	}
	if !strings.Contains(msg.Prefix, "trash") {
		ylogger.Zero.Info().Msg("prefix doesn't contain trash aborted")
		return nil
	}
	trashRetention := time.Hour * 24 * time.Duration(dh.Cnf.TrashRetentionDays)
	filtered := fileList[:0]
	for _, file := range fileList {
		if strings.Contains(file.Path, "trash") && file.LastMod.Add(trashRetention).Unix() < time.Now().Unix() {
			filtered = append(filtered, file)
		}
	}
	fileList = filtered
	for retryCount := 0; len(fileList) > 0 && retryCount < 10; retryCount++ {
		fileList, err = dh.garbageTrashParallel(bucket, fileList)
		if err == nil {
			break
		}
	}

	if len(fileList) > 0 {
		ylogger.Zero.Error().Str("bucket", bucket).Int("failed files count", len(fileList)).Msg("some files were not deleted")
		ylogger.Zero.Error().Str("bucket", bucket).Any("failed files", fileList).Msg("failed to delete some files")
		return errors.Wrap(err, "failed to delete some files")
	}

	for key, uploadId := range uploads {
		if err := dh.StorageInterractor.AbortMultipartUpload(bucket, key, uploadId); err != nil {
			return err
		}
	}

	return nil
}

func (dh *BasicGarbageMgr) HandleDelete2Prefix(msg message.Delete2Message) error {
	for _, b := range dh.StorageInterractor.ListBuckets() {
		if err := dh.DeletePrefixInBucket(b, msg); err != nil {
			return err
		}
	}
	return nil
}
func (dh *BasicGarbageMgr) HandleDeleteFile(msg message.DeleteMessage) error {
	if !msg.Confirm {
		return nil
	}
	for _, b := range dh.StorageInterractor.ListBuckets() {
		err := dh.StorageInterractor.DeleteObject(b, msg.Name)
		if err != nil {
			ylogger.Zero.Error().AnErr("err", err).Str("name", msg.Name).Msg("failed to delete file")
			return errors.Wrap(err, "failed to delete file")
		}
	}
	return nil
}

func (dh *BasicGarbageMgr) ListGarbageFiles(bucket string, msg message.DeleteMessage) ([]*object.ObjectInfo, error) {
	procStartTime := time.Now()

	// Get first backup lsn
	var firstBackupLSN uint64
	var err error

	if dh.Cnf.CheckBackup {
		firstBackupLSN, err = dh.BackupInterractor.GetFirstLSN(msg.Segnum)
		if err != nil {
			ylogger.Zero.Error().AnErr("err", err).Msg("failed to get first lsn") //return or just assume there are no backups?
			return nil, err
		}
		ylogger.Zero.Info().Uint64("lsn", firstBackupLSN).Msg("first backup LSN")
	} else {
		firstBackupLSN = ^uint64(0)
		ylogger.Zero.Info().Uint64("lsn", firstBackupLSN).Msg("omit first backup LSN")
	}

	// List files in storage
	ylogger.Zero.Info().Str("path", msg.Name).Msg("listing prefix")
	objectMetas, err := dh.StorageInterractor.ListBucketPath(bucket, msg.Name, true)
	if err != nil {
		return nil, errors.Wrap(err, "could not list objects")
	}
	ylogger.Zero.Info().Int("amount", len(objectMetas)).Msg("objects count")

	vi, ei, err := dh.DbInterractor.GetVirtualExpireIndexes(msg.Port)
	if err != nil {
		ylogger.Zero.Error().AnErr("err", err).Msg("failed to get indexes")
		return nil, errors.Wrap(err, "could not get virtual and expire indexes")
	}
	ylogger.Zero.Info().Int("virtual", len(vi)).Int("expire", len(ei)).Msg("received virtual index and expire index")

	filesToDelete := make([]*object.ObjectInfo, 0)
	for i := range objectMetas {
		reworkedName := objectMetas[i].Path
		ylogger.Zero.Debug().Str("reworked name", reworkedName).Msg("lookup chunk")

		if vi[reworkedName] {
			continue
		}

		// Never delete a file that was created/modified at or after the
		// moment this vacuum procedure started listing storage. Such a file
		// could not have been accounted for by the virtual/expire index
		// snapshot taken above.
		if objectMetas[i].LastMod.After(procStartTime) {
			ylogger.Zero.Debug().Str("file", objectMetas[i].Path).
				Time("last modified", objectMetas[i].LastMod).
				Time("proc start", procStartTime).
				Msg("file was created after vacuum procedure started, skipping")
			continue
		}

		lsn, ok := ei[reworkedName]
		ylogger.Zero.Debug().Uint64("lsn", lsn).Uint64("backup lsn", firstBackupLSN).Str("path", objectMetas[i].Path).Msg("comparing lsn")
		if lsn < firstBackupLSN || !ok {
			ylogger.Zero.Debug().Str("file", objectMetas[i].Path).
				Bool("file in expire index", ok).
				Bool("lsn is less than in first backup", lsn < firstBackupLSN).
				Msg("file does not persist in virtual index, nor needed for PITR, so will be deleted")
			filesToDelete = append(filesToDelete, objectMetas[i])
		}
	}

	ylogger.Zero.Info().Int("amount", len(filesToDelete)).Msg("files will be deleted")

	return filesToDelete, nil
}
