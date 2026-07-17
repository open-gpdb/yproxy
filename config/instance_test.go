package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestConfig(t *testing.T, filename string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	return path
}

func TestReadInstanceConfigUsesDefaultsWhenUnset(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "socket_path: /tmp/yproxy.sock\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	assertDefaultStorage(t, cfg.StorageCnf)
	assertDefaultBackupStorage(t, cfg.BackupStorageCnf)
	assertDefaultVacuum(t, cfg.VacuumCnf)
	if cfg.StatPort != DefaultStatPort {
		t.Fatalf("expected default stat port %v, got %v", DefaultStatPort, cfg.StatPort)
	}
	if cfg.PsqlPort != DefaultPsqlPort {
		t.Fatalf("expected default psql port %v, got %v", DefaultPsqlPort, cfg.PsqlPort)
	}
	if cfg.MetricsPort != DefaultMetricsPort {
		t.Fatalf("expected default metrics port %v, got %v", DefaultMetricsPort, cfg.MetricsPort)
	}
}

func TestReadInstanceConfigPreservesExplicitZeroPortsYAML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "stat_port: 0\npsql_port: 0\nmetrics_port: 0\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.StatPort != 0 {
		t.Fatalf("expected explicit zero stat port, got %v", cfg.StatPort)
	}
	if cfg.PsqlPort != 0 {
		t.Fatalf("expected explicit zero psql port, got %v", cfg.PsqlPort)
	}
	if cfg.MetricsPort != 0 {
		t.Fatalf("expected explicit zero metrics port, got %v", cfg.MetricsPort)
	}
}

func TestReadInstanceConfigPreservesExplicitZeroStorageYAML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "storage:\n  storage_concurrency: 0\n  copy_storage_concurrency: 0\n  storage_rate_limit: 0\n  storage_endpoint_source_scheme: \"\"\nbackup_storage:\n  storage_concurrency: 0\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.StorageCnf.StorageConcurrency != 0 {
		t.Fatalf("expected explicit zero storage concurrency, got %v", cfg.StorageCnf.StorageConcurrency)
	}
	if cfg.StorageCnf.CopyStorageConcurrency != 0 {
		t.Fatalf("expected explicit zero copy storage concurrency, got %v", cfg.StorageCnf.CopyStorageConcurrency)
	}
	if cfg.StorageCnf.StorageRateLimit != 0 {
		t.Fatalf("expected explicit zero storage rate limit, got %v", cfg.StorageCnf.StorageRateLimit)
	}
	if cfg.StorageCnf.EndpointSourceScheme != "" {
		t.Fatalf("expected explicit empty endpoint source scheme, got %q", cfg.StorageCnf.EndpointSourceScheme)
	}
	if cfg.BackupStorageCnf.StorageConcurrency != 0 {
		t.Fatalf("expected explicit zero backup storage concurrency, got %v", cfg.BackupStorageCnf.StorageConcurrency)
	}
}

func TestReadInstanceConfigPreservesExplicitZeroProtectionWindowYAML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "vacuum:\n  protection_window: 0s\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.VacuumCnf.ProtectionWindow != 0 {
		t.Fatalf("expected explicit zero protection window, got %v", cfg.VacuumCnf.ProtectionWindow)
	}
}

func TestReadInstanceConfigPreservesExplicitProtectionWindowYAML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "vacuum:\n  protection_window: 1h\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.VacuumCnf.ProtectionWindow != time.Hour {
		t.Fatalf("expected explicit protection window %v, got %v", time.Hour, cfg.VacuumCnf.ProtectionWindow)
	}
}

func TestReadInstanceConfigPreservesExplicitFalseCheckBackupYAML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "vacuum:\n  check_backup: false\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.VacuumCnf.CheckBackup {
		t.Fatal("expected explicit false check_backup to be preserved")
	}
}

func TestReadInstanceConfigPreservesExplicitZeroProtectionWindowJSON(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.json", `{"vacuum":{"protection_window":0}}`))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.VacuumCnf.ProtectionWindow != 0 {
		t.Fatalf("expected explicit zero protection window, got %v", cfg.VacuumCnf.ProtectionWindow)
	}
}

func TestReadInstanceConfigPreservesExplicitFalseCheckBackupTOML(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.toml", "[vacuum]\ncheck_backup = false\nprotection_window = \"0s\"\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if cfg.VacuumCnf.CheckBackup {
		t.Fatal("expected explicit false check_backup to be preserved")
	}
	if cfg.VacuumCnf.ProtectionWindow != 0 {
		t.Fatalf("expected explicit zero protection window, got %v", cfg.VacuumCnf.ProtectionWindow)
	}
}

func assertDefaultStorage(t *testing.T, storage Storage) {
	t.Helper()

	if storage.StorageType != "s3" {
		t.Fatalf("expected default storage type %q, got %q", "s3", storage.StorageType)
	}
	if storage.StorageConcurrency != DefaultStorageConcurrency {
		t.Fatalf("expected default storage concurrency %v, got %v", DefaultStorageConcurrency, storage.StorageConcurrency)
	}
	if storage.CopyStorageConcurrency != DefaultCopyStorageConcurrency {
		t.Fatalf("expected default copy storage concurrency %v, got %v", DefaultCopyStorageConcurrency, storage.CopyStorageConcurrency)
	}
	if storage.StorageRateLimit != DefaultStorageRateLimit {
		t.Fatalf("expected default storage rate limit %v, got %v", DefaultStorageRateLimit, storage.StorageRateLimit)
	}
	if storage.EndpointSourceScheme != DefaultEndpointSourceScheme {
		t.Fatalf("expected default endpoint source scheme %q, got %q", DefaultEndpointSourceScheme, storage.EndpointSourceScheme)
	}
}

func assertDefaultBackupStorage(t *testing.T, storage Storage) {
	t.Helper()

	if storage.StorageType != "s3" {
		t.Fatalf("expected default backup storage type %q, got %q", "s3", storage.StorageType)
	}
	if storage.StorageConcurrency != DefaultStorageConcurrency {
		t.Fatalf("expected default backup storage concurrency %v, got %v", DefaultStorageConcurrency, storage.StorageConcurrency)
	}
	if storage.CopyStorageConcurrency != 0 {
		t.Fatalf("expected default backup copy storage concurrency %v, got %v", 0, storage.CopyStorageConcurrency)
	}
	if storage.StorageRateLimit != 0 {
		t.Fatalf("expected default backup storage rate limit %v, got %v", 0, storage.StorageRateLimit)
	}
	if storage.EndpointSourceScheme != "" {
		t.Fatalf("expected default backup endpoint source scheme %q, got %q", "", storage.EndpointSourceScheme)
	}
}

func assertDefaultVacuum(t *testing.T, vacuum Vacuum) {
	t.Helper()

	if vacuum.CheckBackup != DefaultCheckBackup {
		t.Fatalf("expected default check backup %v, got %v", DefaultCheckBackup, vacuum.CheckBackup)
	}
	if vacuum.FileChunkPerSec != DefaultFileChunkPerSec {
		t.Fatalf("expected default file chunk per sec %v, got %v", DefaultFileChunkPerSec, vacuum.FileChunkPerSec)
	}
	if vacuum.TrashRetentionDays != DefaultTrashRetentionDays {
		t.Fatalf("expected default trash retention days %v, got %v", DefaultTrashRetentionDays, vacuum.TrashRetentionDays)
	}
	if vacuum.TrashDeleteWorkers != DefaultTrashDeleteWorkers {
		t.Fatalf("expected default trash delete workers %v, got %v", DefaultTrashDeleteWorkers, vacuum.TrashDeleteWorkers)
	}
	if vacuum.ProtectionWindow != DefaultProtectionWindow {
		t.Fatalf("expected default protection window %v, got %v", DefaultProtectionWindow, vacuum.ProtectionWindow)
	}
}
