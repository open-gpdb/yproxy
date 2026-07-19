package config

import "time"

const (
	DefaultCheckBackup        = true
	DefaultFileChunkPerSec    = 1000
	DefaultTrashRetentionDays = 7
	DefaultTrashDeleteWorkers = 1
	DefaultProtectionWindow   = 24 * time.Hour
)

type Vacuum struct {
	CheckBackup        bool          `json:"check_backup" toml:"check_backup" yaml:"check_backup"`
	FileChunkPerSec    int           `json:"file_chunk_per_sec" toml:"file_chunk_per_sec" yaml:"file_chunk_per_sec"`
	TrashRetentionDays int           `json:"trash_retention_days" toml:"trash_retention_days" yaml:"trash_retention_days"`
	TrashDeleteWorkers int           `json:"trash_delete_workers" toml:"trash_delete_workers" yaml:"trash_delete_workers"`
	ProtectionWindow   time.Duration `json:"protection_window" toml:"protection_window" yaml:"protection_window"`
}

type VacuumOption func(*Vacuum)

func WithCheckBackup(checkBackup bool) VacuumOption {
	return func(v *Vacuum) {
		v.CheckBackup = checkBackup
	}
}

func WithFileChunkPerSec(fileChunkPerSec int) VacuumOption {
	return func(v *Vacuum) {
		v.FileChunkPerSec = fileChunkPerSec
	}
}

func WithTrashRetentionDays(trashRetentionDays int) VacuumOption {
	return func(v *Vacuum) {
		v.TrashRetentionDays = trashRetentionDays
	}
}

func WithTrashDeleteWorkers(trashDeleteWorkers int) VacuumOption {
	return func(v *Vacuum) {
		v.TrashDeleteWorkers = trashDeleteWorkers
	}
}

func WithProtectionWindow(protectionWindow time.Duration) VacuumOption {
	return func(v *Vacuum) {
		v.ProtectionWindow = protectionWindow
	}
}

func BuildVacuum(opts ...VacuumOption) *Vacuum {
	v := &Vacuum{}

	ApplyVacuumOptions(v,
		WithCheckBackup(DefaultCheckBackup),
		WithFileChunkPerSec(DefaultFileChunkPerSec),
		WithTrashRetentionDays(DefaultTrashRetentionDays),
		WithTrashDeleteWorkers(DefaultTrashDeleteWorkers),
		WithProtectionWindow(DefaultProtectionWindow),
	)
	ApplyVacuumOptions(v, opts...)

	return v
}

func ApplyVacuumOptions(v *Vacuum, opts ...VacuumOption) {
	for _, opt := range opts {
		opt(v)
	}
}
