package storage

import "github.com/open-gpdb/yproxy/pkg/settings"

func ResolveStorageSetting(settings []settings.StorageSettings, name, defaultVal string) string {

	for _, s := range settings {
		if s.Name == name {
			return s.Value
		}
	}

	return defaultVal
}
