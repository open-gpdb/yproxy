package config

import "time"

type Vacuum struct {
	CheckBackup        bool          `json:"check_backup" toml:"check_backup" yaml:"check_backup"`
	FileChunkPerSec    int           `json:"file_chunk_per_sec" toml:"file_chunk_per_sec" yaml:"file_chunk_per_sec"`
	TrashRetentionDays int           `json:"trash_retention_days" toml:"trash_retention_days" yaml:"trash_retention_days"`
	TrashDeleteWorkers int           `json:"trash_delete_workers" toml:"trash_delete_workers" yaml:"trash_delete_workers"`
	ProtectionWindow   time.Duration `json:"protection_window" toml:"protection_window" yaml:"protection_window"`
}
