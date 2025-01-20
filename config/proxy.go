package config

type Proxy struct {
	ConsolePort string `json:"console_port" toml:"console_port" yaml:"console_port"`

	BucketCachePath string `json:"bucket_cache_path" toml:"bucket_cache_path" yaml:"bucket_cache_path"`
}
