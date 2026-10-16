package storage

import (
	"errors"

	"github.com/yezzey-gp/aws-sdk-go/aws/awserr"
	"github.com/yezzey-gp/yproxy/pkg/settings"
)

func ResolveStorageSetting(settings []settings.StorageSettings, name, defaultVal string) string {

	for _, s := range settings {
		if s.Name == name {
			return s.Value
		}
	}

	return defaultVal
}

/*
 * Storage returns 507 for objects that do not exist.
 * Retrying such requests is pointless: the object will not appear.
 */
const nonRetryableStatusCode = 507

func IsNonRetryableError(err error) bool {
	var reqErr awserr.RequestFailure
	if errors.As(err, &reqErr) {
		return reqErr.StatusCode() == nonRetryableStatusCode
	}
	return false
}
