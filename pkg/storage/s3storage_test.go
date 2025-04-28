package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/object"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)

	f, err := os.CreateTemp("", "*")
	assert.NoError(err)
	defer func() { _ = os.Remove(f.Name()) }()
	defer func() { _ = f.Close() }()

	config.InstanceConfig().ProxyCnf.BucketCachePath = f.Name()

	s1 := config.Storage{StorageBucket: "some_bucket_1"}
	s2 := config.Storage{StorageBucket: "some_bucket_2"}

	abcObjects := []*object.ObjectInfo{{Path: "/abc1"}, {Path: "/abc2"}}
	allObjects := append(abcObjects, &object.ObjectInfo{Path: "/def1"})
	err = putInCache(s1.ID(), allObjects)
	assert.NoError(err)

	err = putInCache(s2.ID(), allObjects)
	assert.NoError(err)

	objects, err := readCache(s1, "")
	assert.NoError(err)
	assert.Equal(allObjects, objects)

	objects, err = readCache(s2, "abc")
	assert.NoError(err)
	assert.Equal(abcObjects, objects)
}
