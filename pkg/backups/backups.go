package backups

type BackupLSN struct {
	Lsn uint64 `json:"LSN"`
}

//go:generate mockgen -destination=pkg/mock/backups.go -package=mock
type BackupInterractor interface {
	GetFirstLSN(seg uint64) (uint64, error)
}
