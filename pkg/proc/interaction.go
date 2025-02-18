package proc

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/backups"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	"github.com/yezzey-gp/yproxy/pkg/database"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/proc/yio"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
	"golang.org/x/sync/semaphore"
)

func ProcessCatExtended(
	s storage.StorageInteractor,
	pr *ProtoReader,
	name string,
	decrypt bool, kek bool, startOffset uint64, settings []settings.StorageSettings, cr crypt.Crypter, ycl client.YproxyClient) error {

	ycl.SetExternalFilePath(name)

	yr := yio.NewYRetryReader(yio.NewRestartReader(s, name, settings), ycl)

	var contentReader io.Reader
	contentReader = yr
	defer yr.Close()
	var err error

	if decrypt {
		if cr == nil {
			err := fmt.Errorf("failed to decrypt object, decrypter not configured")
			_ = ycl.ReplyError(err, "cat failed")
			return err
		}
		ylogger.Zero.Debug().Str("object-path", name).Msg("decrypt object")
		contentReader, err = cr.Decrypt(yr)
		if err != nil {
			_ = ycl.ReplyError(err, "failed to decrypt object")

			return err
		}
	}

	if kek {
		err := fmt.Errorf("KEK is currently unsupported")
		_ = ycl.ReplyError(err, "cat failed")
		return err
	}

	if startOffset != 0 {
		io.CopyN(io.Discard, contentReader, int64(startOffset))
	}

	n, err := io.Copy(ycl.GetRW(), contentReader)
	if err != nil {
		_ = ycl.ReplyError(err, "copy failed to complete")
	}
	ylogger.Zero.Debug().Int64("copied bytes", n).Msg("decrypt object")

	return nil
}

func ProcessPutExtended(
	s storage.StorageInteractor,
	pr *ProtoReader,
	name string,
	encrypt bool, settings []settings.StorageSettings, cr crypt.Crypter, ycl client.YproxyClient,
	replyKV bool) error {

	ycl.SetExternalFilePath(name)

	var w io.WriteCloser
	r, w := io.Pipe()

	w = yio.NewYproxyWriter(w, ycl)

	defer r.Close()
	defer w.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		var ww io.WriteCloser = w
		if encrypt {
			if cr == nil {
				_ = ycl.ReplyError(fmt.Errorf("failed to encrypt, crypter not configured"), "connection aborted")

				return
			}

			var err error
			ww, err = cr.Encrypt(w)
			if err != nil {
				_ = ycl.ReplyError(err, "failed to encrypt")
				return
			}
		} else {
			ylogger.Zero.Debug().Str("path", name).Msg("omit encryption for upload chunks")
		}

		defer func() {
			if err := ww.Close(); err != nil {
				_ = ycl.ReplyError(err, "failed to close connection")
				return
			}

			if encrypt {
				if err := w.Close(); err != nil {
					ylogger.Zero.Error().Err(err).Msg("failed to close connection")
					return
				}
			}

			ylogger.Zero.Debug().Msg("closing msg writer")
		}()

		for {
			tp, body, err := pr.ReadPacket()
			if err != nil {
				_ = ycl.ReplyError(err, "failed to read chunk of data")
				return
			}

			ylogger.Zero.Debug().Str("msg-type", tp.String()).Msg("received client request")

			switch tp {
			case message.MessageTypeCopyData:
				msg := message.CopyDataMessage{}
				msg.Decode(body)
				if n, err := ww.Write(msg.Data); err != nil {
					_ = ycl.ReplyError(err, "failed to write copy data")

					return
				} else if n != int(msg.Sz) {

					_ = ycl.ReplyError(fmt.Errorf("unfull write"), "failed to complete request")

					return
				}
			case message.MessageTypeCommandComplete:
				msg := message.CommandCompleteMessage{}
				msg.Decode(body)
				return
			default:
				return
			}
		}
	}()

	for _, s := range settings {
		ylogger.Zero.Debug().Bool("encypt", encrypt).Str("name", name).Str("name", s.Name).Str("value", s.Value).Msg("offloading setting")
	}

	/* Should go after reader dispatch! */
	if err := s.PutFileToDest(name, r, settings); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")

		return err
	}

	wg.Wait()

	if replyKV {
		if _, err := ycl.GetRW().Write(message.NewPutCompleteMessage(1).Encode()); err != nil {
			_ = ycl.ReplyError(err, "failed to upload")

			return err
		}
	}

	if _, err := ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")

		return err
	}

	return nil
}
func ProcessListExtended(msg message.ListMessage, s storage.StorageInteractor, cr crypt.Crypter, ycl client.YproxyClient, cnf *config.Vacuum) error {
	ycl.SetExternalFilePath(msg.Prefix)

	objectMetas, err := s.ListPath(msg.Prefix, true)
	if err != nil {
		_ = ycl.ReplyError(fmt.Errorf("could not list objects: %s", err), "failed to complete request")

		return nil
	}

	const chunkSize = 1000

	for i := 0; i < len(objectMetas); i += chunkSize {
		_, err = ycl.GetRW().Write(message.NewObjectMetaMessage(objectMetas[i:min(i+chunkSize, len(objectMetas))]).Encode())
		if err != nil {
			_ = ycl.ReplyError(err, "failed to upload")

			return nil
		}

	}

	_, err = ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode())

	if err != nil {
		_ = ycl.ReplyError(err, "failed to upload")

		return err
	}
	return nil
}
func ProcessCopyExtended(msg message.CopyMessage, s storage.StorageInteractor, cr crypt.Crypter, ycl client.YproxyClient) error {

	ycl.SetExternalFilePath(msg.Name)

	// get config for old bucket
	instanceCnf, err := config.ReadInstanceConfig(msg.OldCfgPath)
	if err != nil {
		_ = ycl.ReplyError(fmt.Errorf("could not read old config: %s", err), "failed to compelete request")
		return nil
	}
	config.EmbedDefaults(&instanceCnf)
	oldStorage, err := storage.NewStorage(&instanceCnf.StorageCnf)
	if err != nil {
		return err
	}
	ylogger.Zero.Info().Interface("cnf", instanceCnf).Msg("loaded new config")

	objectMetas, _, err := ListFilesToCopy(msg.Name, msg.Port, instanceCnf.StorageCnf, oldStorage, s)
	if err != nil {
		return err
	}

	if !msg.Confirm {
		ylogger.Zero.Info().Msg("It was a dry-run, nothing was copied")
		return nil
	}

	var my sync.Mutex

	var failed []*object.ObjectInfo
	retryCount := 0
	for len(objectMetas) > 0 && retryCount < 10 {
		retryCount++

		sem := semaphore.NewWeighted(200)

		wg := sync.WaitGroup{}

		for i := range len(objectMetas) {
			path := strings.TrimPrefix(objectMetas[i].Path, instanceCnf.StorageCnf.StoragePrefix)

			sem.Acquire(context.TODO(), 1)
			wg.Add(1)

			go func(i int) {
				defer sem.Release(1)
				defer wg.Done()

				ylogger.Zero.Info().Int("index", i).Str("object path", objectMetas[i].Path).Int64("object size", objectMetas[i].Size).Msg("copying...")
				/* get reader */
				readerFromOldBucket := yio.NewYRetryReader(yio.NewRestartReader(oldStorage, path, nil), ycl)
				var fromReader io.Reader
				fromReader = readerFromOldBucket
				defer readerFromOldBucket.Close()

				if msg.Decrypt {
					oldCr, err := crypt.NewCrypto(&instanceCnf.CryptoCnf)
					if err != nil {
						ylogger.Zero.Error().Err(err).Msg("failed to configure decrypter")
						my.Lock()
						failed = append(failed, objectMetas[i])
						my.Unlock()
						return
					}
					fromReader, err = oldCr.Decrypt(readerFromOldBucket)
					if err != nil {
						ylogger.Zero.Error().Err(err).Msg("failed to decrypt object")
						my.Lock()
						failed = append(failed, objectMetas[i])
						my.Unlock()
						return
					}
				}

				/* re-encrypt */
				readerEncrypt, writerEncrypt := io.Pipe()

				go func() {
					defer func() {
						if err := writerEncrypt.Close(); err != nil {
							ylogger.Zero.Warn().Err(err).Msg("failed to close writer")
						}
					}()

					var writerToNewBucket io.WriteCloser = writerEncrypt

					if msg.Encrypt {
						var err error
						writerToNewBucket, err = cr.Encrypt(writerEncrypt)
						if err != nil {
							ylogger.Zero.Error().Err(err).Msg("failed to encrypt object")
							my.Lock()
							failed = append(failed, objectMetas[i])
							my.Unlock()
							return
						}
					}

					if _, err := io.Copy(writerToNewBucket, fromReader); err != nil {
						ylogger.Zero.Error().Str("path", path).Err(err).Msg("failed to copy data")
						my.Lock()
						failed = append(failed, objectMetas[i])
						my.Unlock()
						return
					}

					if err := writerToNewBucket.Close(); err != nil {
						ylogger.Zero.Error().Str("path", path).Err(err).Msg("failed to close writer")
						my.Lock()
						failed = append(failed, objectMetas[i])
						my.Unlock()
						return
					}
				}()

				//write file
				err = s.PutFileToDest(path, readerEncrypt, nil)
				if err != nil {
					ylogger.Zero.Error().Err(err).Msg("failed to upload file")
					my.Lock()
					failed = append(failed, objectMetas[i])
					my.Unlock()
					return
				}
			}(i)
		}
		wg.Wait()
		objectMetas = failed
		ylogger.Zero.Info().Int("count", len(objectMetas)).Msg("failed files count")
		failed = make([]*object.ObjectInfo, 0)
	}

	if len(objectMetas) > 0 {
		ylogger.Zero.Info().Int("count", len(objectMetas)).Msg("failed files count")
		fmt.Printf("failed files: %v\n", objectMetas)
		ylogger.Zero.Error().Int("failed files count", len(objectMetas)).Msg("failed to upload some files")
		ylogger.Zero.Error().Any("failed files", objectMetas).Msg("failed to upload some files")

		err := fmt.Errorf("failed to copy some files")

		_ = ycl.ReplyError(err, "failed files")
		return err
	}

	if _, err = ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")
		return err
	}
	ylogger.Zero.Info().Msg("Copy finished successfully")
	return nil
}

func ProcessDeleteExtended(msg message.DeleteMessage, s storage.StorageInteractor, bs storage.StorageInteractor, ycl client.YproxyClient, cnf *config.Vacuum) error {
	ycl.SetExternalFilePath(msg.Name)

	dbInterractor := &database.DatabaseHandler{}
	backupHandler := &backups.StorageBackupInteractor{Storage: bs}

	var dh = &BasicGarbageMgr{
		StorageInterractor: s,
		DbInterractor:      dbInterractor,
		BackupInterractor:  backupHandler,
		Cnf:                cnf,
	}

	if msg.Garbage {
		ylogger.Zero.Debug().
			Str("Name", msg.Name).
			Uint64("port", msg.Port).
			Uint64("segment", msg.Segnum).
			Bool("confirm", msg.Confirm).Msg("requested to perform external storage VACUUM")
	} else {
		ylogger.Zero.Debug().
			Str("Name", msg.Name).
			Uint64("port", msg.Port).
			Uint64("segment", msg.Segnum).
			Bool("confirm", msg.Confirm).Msg("requested to remove external chunk")
	}

	if msg.Garbage {
		err := dh.HandleDeleteGarbage(msg)
		if err != nil {
			_ = ycl.ReplyError(err, "failed to finish operation")
			return err
		}
	} else {
		err := dh.HandleDeleteFile(msg)
		if err != nil {
			_ = ycl.ReplyError(err, "failed to finish operation")
			return err
		}
	}

	if _, err := ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")
		return err
	}

	if msg.Garbage {
		if !msg.Confirm {
			ylogger.Zero.Warn().Msg("It was a dry-run, nothing was deleted")
		} else {
			ylogger.Zero.Info().Msg("Deleted garbage successfully")
		}
	} else {
		ylogger.Zero.Info().Msg("Deleted chunk successfully")
	}

	return nil
}

func ProcessUntrashify(msg message.UntrashifyMessage, s storage.StorageInteractor, bs storage.StorageInteractor, ycl client.YproxyClient) error {
	ycl.SetExternalFilePath(msg.Name)

	dbInterractor := &database.DatabaseHandler{}
	backupHandler := &backups.StorageBackupInteractor{Storage: bs}

	var dh = &BasicGarbageMgr{
		StorageInterractor: s,
		DbInterractor:      dbInterractor,
		BackupInterractor:  backupHandler,
		Cnf:                &config.Vacuum{},
	}

	ylogger.Zero.Debug().
		Str("Name", msg.Name).
		Uint64("port", msg.Port).
		Uint64("segment", msg.Segnum).
		Bool("confirm", msg.Confirm).Msg("requested to perform untrashify")

	if err := dh.HandleUntrasifyFile(msg); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")
		return err
	}

	if _, err := ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode()); err != nil {
		_ = ycl.ReplyError(err, "failed to upload")
		return err
	}

	if !msg.Confirm {
		ylogger.Zero.Warn().Msg("It was a dry-run, nothing was deleted")
	} else {
		ylogger.Zero.Info().Msg("Untrashify garbage successfully")
	}

	return nil
}

func ProcConn(s storage.StorageInteractor, bs storage.StorageInteractor, cr crypt.Crypter, ycl client.YproxyClient, cnf *config.Vacuum) error {

	defer func() {
		_ = ycl.Close()
	}()

	pr := NewProtoReader(ycl)
	tp, body, err := pr.ReadPacket()
	if err != nil {
		_ = ycl.ReplyError(err, "failed to read request packet")
		return err
	}

	ylogger.Zero.Debug().Str("msg-type", tp.String()).Msg("received client request")

	ycl.SetOPType(tp)

	switch tp {
	case message.MessageTypeCat:

		// omit first byte
		msg := message.CatMessage{}
		msg.Decode(body)

		if err := ProcessCatExtended(s, pr, msg.Name, msg.Decrypt, false, msg.StartOffset, nil, cr, ycl); err != nil {
			return err
		}

	case message.MessageTypeCatV2:
		// omit first byte
		msg := message.CatMessageV2{}
		msg.Decode(body)

		if err := ProcessCatExtended(s, pr, msg.Name, msg.Decrypt, msg.KEK, msg.StartOffset, msg.Settings, cr, ycl); err != nil {
			return err
		}

	case message.MessageTypePut:

		msg := message.PutMessage{}
		msg.Decode(body)

		if err := ProcessPutExtended(s, pr, msg.Name, msg.Encrypt, nil, cr, ycl, false); err != nil {
			return err
		}

	case message.MessageTypePutV2:

		msg := message.PutMessageV2{}
		msg.Decode(body)

		if err := ProcessPutExtended(s, pr, msg.Name, msg.Encrypt, msg.Settings, cr, ycl, false); err != nil {
			return err
		}

	case message.MessageTypePutV3:
		msg := message.PutMessageV3{}
		msg.Decode(body)

		if err := ProcessPutExtended(s, pr, msg.Name, msg.Encrypt, msg.Settings, cr, ycl, true); err != nil {
			return err
		}

	case message.MessageTypeList:
		msg := message.ListMessage{}
		msg.Decode(body)

		err := ProcessListExtended(msg, s, cr, ycl, cnf)
		if err != nil {
			return err
		}
	case message.MessageTypeCopy:
		msg := message.CopyMessage{}
		msg.Decode(body)

		err := ProcessCopyExtended(msg, s, cr, ycl)
		if err != nil {
			return err
		}
	case message.MessageTypeDelete:
		//recieve message
		msg := message.DeleteMessage{}
		msg.Decode(body)
		err := ProcessDeleteExtended(msg, s, bs, ycl, cnf)
		if err != nil {
			return err
		}

	case message.MessageTypeUntrashify:
		//recieve message
		msg := message.UntrashifyMessage{}
		msg.Decode(body)
		err := ProcessUntrashify(msg, s, bs, ycl)
		if err != nil {
			return err
		}
	case message.MessageTypeGool:
		return ProcMotion(s, cr, ycl)

	default:
		ylogger.Zero.Error().Any("type", tp).Msg("unknown message type")
		_ = ycl.ReplyError(nil, "wrong request type")

		return nil
	}

	return nil
}

func ProcMotion(s storage.StorageInteractor, cr crypt.Crypter, ycl client.YproxyClient) error {

	defer func() {
		_ = ycl.Close()
	}()

	pr := NewProtoReader(ycl)
	tp, body, err := pr.ReadPacket()
	if err != nil {
		_ = ycl.ReplyError(err, "failed to read request packet")
		return err
	}

	ylogger.Zero.Debug().Str("msg-type", tp.String()).Msg("received client request")

	ycl.SetOPType(tp)

	msg := message.GoolMessage{}
	msg.Decode(body)

	ylogger.Zero.Info().Msg("received client gool succ")

	_, err = ycl.GetRW().Write(message.NewReadyForQueryMessage().Encode())
	if err != nil {
		_ = ycl.ReplyError(err, "failed to gool")
	}
	return nil
}

func ListFilesToCopy(prefix string, port uint64, cfg config.Storage, src storage.StorageLister, dst storage.StorageLister) ([]*object.ObjectInfo, []*object.ObjectInfo, error) {
	objectMetas, err := src.ListPath(prefix, true)
	if err != nil {
		return nil, nil, err
	}

	dbInterractor := &database.DatabaseHandler{}
	vi, _, err := dbInterractor.GetVirtualExpireIndexes(port)
	if err != nil {
		return nil, nil, err
	}

	copied, err := dst.ListPath(prefix, false)
	if err != nil {
		return nil, nil, err
	}
	copiedSizes := make(map[string]int64)
	for _, c := range copied {
		copiedSizes[c.Path] = c.Size
	}

	toCopy := []*object.ObjectInfo{}
	skipped := []*object.ObjectInfo{}

	for i := range len(objectMetas) {
		path := strings.TrimPrefix(objectMetas[i].Path, cfg.StoragePrefix)
		reworked := path
		if _, ok := vi[reworked]; !ok {
			ylogger.Zero.Info().Int("index", i).Str("object path", objectMetas[i].Path).Msg("not in virtual index, skipping...")
			skipped = append(skipped, objectMetas[i])
			continue
		}
		if sz, ok := copiedSizes[objectMetas[i].Path]; ok {
			ylogger.Zero.Info().Int("index", i).Str("object path", objectMetas[i].Path).Int64("object size", objectMetas[i].Size).Int64("copeid size", sz).Msg("already copied, skipping...")
			skipped = append(skipped, objectMetas[i])
			continue
		}

		ylogger.Zero.Info().Str("object path", objectMetas[i].Path).Int64("object size", objectMetas[i].Size).Msg("will be copied")

		toCopy = append(toCopy, objectMetas[i])
	}

	return toCopy, skipped, nil
}
