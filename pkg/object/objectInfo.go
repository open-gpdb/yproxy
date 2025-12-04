package object

import "time"

type ObjectInfo struct {
	Path    string
	Size    int64
	LastMod time.Time
}
