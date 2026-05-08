package config

type Vacuum struct {
	CheckBackup     bool `json:"check_backup" toml:"check_backup" yaml:"check_backup"`
	FileChunkPerSec int  `json:"file_chunk_per_sec" toml:"file_chunk_per_sec" yaml:"file_chunk_per_sec"`
}
