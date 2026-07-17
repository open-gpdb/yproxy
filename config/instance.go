package config

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
)

type Instance struct {
	StorageCnf       Storage `json:"storage" toml:"storage" yaml:"storage"`
	BackupStorageCnf Storage `json:"backup_storage" toml:"backup_storage" yaml:"backup_storage"`
	ProxyCnf         Proxy   `json:"proxy" toml:"proxy" yaml:"proxy"`

	CryptoCnf Crypto `json:"crypto" toml:"crypto" yaml:"crypto"`

	VacuumCnf Vacuum `json:"vacuum" toml:"vacuum" yaml:"vacuum"`

	LogPath                string `json:"log_path" toml:"log_path" yaml:"log_path"`
	LogLevel               string `json:"log_level" toml:"log_level" yaml:"log_level"`
	SocketPath             string `json:"socket_path" toml:"socket_path" yaml:"socket_path"`
	StatPort               int    `json:"stat_port" toml:"stat_port" yaml:"stat_port"`
	PsqlPort               int    `json:"psql_port" toml:"psql_port" yaml:"psql_port"`
	InterconnectSocketPath string `json:"interconnect_socket_path" toml:"interconnect_socket_path" yaml:"interconnect_socket_path"`
	DebugPort              int    `json:"debug_port" toml:"debug_port" yaml:"debug_port"`
	DebugMinutes           int    `json:"debug_minutes" toml:"debug_minutes" yaml:"debug_minutes"`
	MetricsPort            int    `json:"metrics_port" toml:"metrics_port" yaml:"metrics_port"`

	YezzeyRestoreParanoid bool `json:"yezzey_restore_paranoid" toml:"yezzey_restore_paranoid" yaml:"yezzey_restore_paranoid"`

	SystemdNotificationsDebug bool `json:"sd_notifications_debug" toml:"sd_notifications_debug" yaml:"sd_notifications_debug"`
	systemdSocketPath         string
}

func (i *Instance) ReadSystemdSocketPath() {
	path := os.Getenv("NOTIFY_SOCKET")
	if path != "" {
		i.systemdSocketPath = path
	}
}

func (i *Instance) GetSystemdSocketPath() string {
	return i.systemdSocketPath
}

var cfgInstance Instance

func InstanceConfig() *Instance {
	return &cfgInstance
}

type InstanceOption func(*Instance)

func WithStorageCnf(storage Storage) InstanceOption {
	return func(i *Instance) {
		i.StorageCnf = storage
	}
}

func WithBackupStorageCnf(storage Storage) InstanceOption {
	return func(i *Instance) {
		i.BackupStorageCnf = storage
	}
}

func WithVacuumCnf(vacuum Vacuum) InstanceOption {
	return func(i *Instance) {
		i.VacuumCnf = vacuum
	}
}

func WithStatPort(statPort int) InstanceOption {
	return func(i *Instance) {
		i.StatPort = statPort
	}
}

func WithPsqlPort(psqlPort int) InstanceOption {
	return func(i *Instance) {
		i.PsqlPort = psqlPort
	}
}

func WithMetricsPort(metricsPort int) InstanceOption {
	return func(i *Instance) {
		i.MetricsPort = metricsPort
	}
}

const (
	DefaultStatPort    = 7432
	DefaultPsqlPort    = 8432
	DefaultMetricsPort = 2112
)

func BuildInstance(opts ...InstanceOption) *Instance {
	i := &Instance{}

	ApplyInstanceOptions(i,
		WithStorageCnf(*BuildStorage()),
		WithBackupStorageCnf(*BuildBackupStorage()),
		WithVacuumCnf(*BuildVacuum()),
		WithStatPort(DefaultStatPort),
		WithPsqlPort(DefaultPsqlPort),
		WithMetricsPort(DefaultMetricsPort),
	)
	ApplyInstanceOptions(i, opts...)

	return i
}

func ApplyInstanceOptions(i *Instance, opts ...InstanceOption) {
	for _, opt := range opts {
		opt(i)
	}
}

func initInstanceConfig(file *os.File, cfgInstance *Instance) error {
	*cfgInstance = *BuildInstance()
	if strings.HasSuffix(file.Name(), ".toml") {
		_, err := toml.NewDecoder(file).Decode(cfgInstance)
		return err
	}
	if strings.HasSuffix(file.Name(), ".yaml") {
		return yaml.NewDecoder(file).Decode(&cfgInstance)
	}
	if strings.HasSuffix(file.Name(), ".json") {
		return json.NewDecoder(file).Decode(&cfgInstance)
	}
	return fmt.Errorf("unknown config format type: %s. Use .toml, .yaml or .json suffix in filename", file.Name())
}

var bootstrapCfgPath = ""

func ReloadInstanceConfig() (*Instance, error) {
	if err := LoadInstanceConfig(bootstrapCfgPath); err != nil {
		return nil, err
	}
	return InstanceConfig(), nil
}

func LoadInstanceConfig(cfgPath string) (err error) {
	if bootstrapCfgPath != "" && bootstrapCfgPath != cfgPath {
		return fmt.Errorf("bootstrap config path already set")
	}
	bootstrapCfgPath = cfgPath
	cfgInstance, err = ReadInstanceConfig(cfgPath)
	if err != nil {
		return
	}

	cfgInstance.ReadSystemdSocketPath()

	configBytes, err := json.MarshalIndent(cfgInstance, "", "  ")
	if err != nil {
		return
	}

	log.Println("Running config:", string(configBytes))
	return
}

func ReadInstanceConfig(cfgPath string) (Instance, error) {
	var cfg Instance
	file, err := os.Open(cfgPath)
	if err != nil {
		return cfg, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("failed to close config file: %v", err)
		}
	}(file)

	if err := initInstanceConfig(file, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
