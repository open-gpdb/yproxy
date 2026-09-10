package proto

import (
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/client"
	"github.com/yezzey-gp/yproxy/pkg/crypt"
	pio "github.com/yezzey-gp/yproxy/pkg/io"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/storage"
)

type ProtoMgr interface {
	ProcessCatExtended(
		s storage.StorageInteractor,
		pr *pio.ProtoReader,
		name string,
		decrypt bool,
		kek bool,
		startOffset uint64,
		settings []settings.StorageSettings,
		cr crypt.Crypter,
		ycl client.YproxyClient) error

	ProcessPutExtended(
		s storage.StorageInteractor,
		pr *pio.ProtoReader,
		name string,
		encrypt bool,
		settings []settings.StorageSettings,
		cr crypt.Crypter,
		ycl client.YproxyClient,
		replyKV bool) error

	ProcessListExtended(prefix string,
		settings []settings.StorageSettings,
		s storage.StorageInteractor,
		cr crypt.Crypter,
		ycl client.YproxyClient,
		cnf *config.Vacuum) error

	ProcessCopyExtended(
		name string,
		oldCfgPath string,
		port uint64,
		confirm,
		encrypt,
		decrypt,
		kEKDecrypt,
		serverSide,
		replyKV bool,
		s storage.StorageInteractor,
		cr crypt.Crypter,
		ycl client.YproxyClient) error

	/* TODO: merge these two */
	ProcessCleanupExtended(
		msg message.CleanupMessage,
		s storage.StorageInteractor,
		bs storage.StorageInteractor,
		ycl client.YproxyClient,
		cnf *config.Vacuum) error

	ProcessDropExtended(
		msg message.DropMessage,
		s storage.StorageInteractor,
		bs storage.StorageInteractor,
		ycl client.YproxyClient,
		cnf *config.Vacuum) error

	ProcessUntrashify(
		msg message.UntrashifyMessage,
		s storage.StorageInteractor,
		bs storage.StorageInteractor,
		ycl client.YproxyClient) error

	/* Delete V3 */
	ProcessCollectObsolete(msg message.CollectObsoleteMessage,
		s storage.StorageInteractor,
		ycl client.YproxyClient) error

	ProcessDeleteObsolete(msg message.DeleteObsoleteMessage,
		s storage.StorageInteractor,
		bs storage.StorageInteractor,
		ycl client.YproxyClient) error
}
