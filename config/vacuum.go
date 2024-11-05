package config

type Vacuum struct {
	CheckBackup bool `json:"check_backup" toml:"check_backup" yaml:"check_backup"`
}
