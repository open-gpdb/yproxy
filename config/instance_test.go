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

func TestReadInstanceConfigUsesDefaultVacuumWhenUnset(t *testing.T) {
	cfg, err := ReadInstanceConfig(writeTestConfig(t, "yproxy.yaml", "socket_path: /tmp/yproxy.sock\n"))
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	assertDefaultVacuum(t, cfg.VacuumCnf)
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
