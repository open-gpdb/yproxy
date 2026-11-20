package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-gpdb/yproxy/pkg/settings"
	"github.com/open-gpdb/yproxy/pkg/storage"
)

func TestResolveSettings(t *testing.T) {

	assert := assert.New(t)

	type tcase struct {
		name     string
		defaultV string
		exp      string
		settings []settings.StorageSettings
	}

	for _, tt := range []tcase{
		{
			"ababa",
			"aboba",
			"aboba",
			nil,
		},
		{
			"ababa",
			"aboba",
			"aboba",
			[]settings.StorageSettings{
				{
					Name:  "djewikdeowp",
					Value: "jdoiwejoidew",
				},
			},
		},

		{
			"ababa",
			"aboba",
			"valval",
			[]settings.StorageSettings{
				{
					Name:  "ababa",
					Value: "valval",
				},
			},
		},
	} {

		assert.Equal(tt.exp, storage.ResolveStorageSetting(tt.settings, tt.name, tt.defaultV))
	}
}
