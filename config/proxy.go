package config

type Proxy struct {
	ConsolePort string `json:"console_port" toml:"console_port" yaml:"console_port"`

	CopySrcBucketCachePath string `json:"copy_src_bucket_cache_path" toml:"copy_src_bucket_cache_path" yaml:"copy_src_bucket_cache_path"`
}
