package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type Instance struct {
	StorageCnf Storage `json:"storage" toml:"storage" yaml:"storage"`
	ProxyCnf   Proxy   `json:"proxy" toml:"proxy" yaml:"proxy"`

	CryptoCnf Crypto `json:"crypto" toml:"crypto" yaml:"crypto"`

	LogPath                string `json:"log_path" toml:"log_path" yaml:"log_path"`
	LogLevel               string `json:"log_level" toml:"log_level" yaml:"log_level"`
	SocketPath             string `json:"socket_path" toml:"socket_path" yaml:"socket_path"`
	StatPort               int    `json:"stat_port" toml:"stat_port" yaml:"stat_port"`
	PsqlPort               int    `json:"psql_port" toml:"psql_port" yaml:"psql_port"`
	InterconnectSocketPath string `json:"interconnect_socket_path" toml:"interconnect_socket_path" yaml:"interconnect_socket_path"`

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

func initInstanceConfig(file *os.File, cfgInstance *Instance) error {
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

const (
	DefaultStorageConcurrency = 100
	DefaultStatPort           = 7432
)

func EmbedDefaults(cfgInstance *Instance) {
	if cfgInstance.StorageCnf.StorageType == "" {
		cfgInstance.StorageCnf.StorageType = "s3"
	}
	if cfgInstance.StorageCnf.StorageConcurrency == 0 {
		cfgInstance.StorageCnf.StorageConcurrency = DefaultStorageConcurrency
	}
	if cfgInstance.StatPort == 0 {
		cfgInstance.StatPort = DefaultStatPort
	}
}

func LoadInstanceConfig(cfgPath string) (err error) {
	cfgInstance, err = ReadInstanceConfig(cfgPath)
	if err != nil {
		return
	}

	cfgInstance.ReadSystemdSocketPath()
	EmbedDefaults(&cfgInstance)

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
