package config

type Crypto struct {
	GPGKeyId   string `json:"gpg_key_id" toml:"gpg_key_id" yaml:"gpg_key_id"`
	GPGKeyPath string `json:"gpg_key_path" toml:"gpg_key_path" yaml:"gpg_key_path"`
}

type Storage struct {
	StorageEndpoint string `json:"storage_endpoint" toml:"storage_endpoint" yaml:"storage_endpoint"`
	AccessKeyId     string `json:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" toml:"secret_access_key" yaml:"secret_access_key"`
	StoragePrefix   string `json:"storage_prefix" toml:"storage_prefix" yaml:"storage_prefix"`
	StorageBucket   string `json:"storage_bucket" toml:"storage_bucket" yaml:"storage_bucket"`

	TablespaceMap map[string]string `json:"tablespace_map" toml:"tablespace_map" yaml:"tablespace_map"`

	// how many concurrrent connection acquire allowed
	StorageConcurrency int64 `json:"storage_concurrency" toml:"storage_concurrency" yaml:"storage_concurrency"`

	StorageRegion string `json:"storage_region" toml:"storage_region" yaml:"storage_region"`

	// File storage default s3. Available: s3, fs
	StorageType string `json:"storage_type" toml:"storage_type" yaml:"storage_type"`

	EndpointSourceHost string `json:"storage_endpoint_source_host" toml:"storage_endpoint_source_host" yaml:"storage_endpoint_source_host"`
	EndpointSourcePort string `json:"storage_endpoint_source_port" toml:"storage_endpoint_source_port" yaml:"storage_endpoint_source_port"`
}
