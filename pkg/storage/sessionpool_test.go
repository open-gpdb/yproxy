package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/aws-sdk-go/aws"
	"github.com/yezzey-gp/aws-sdk-go/service/s3"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/storage"
	"go.nhat.io/httpmock"
)

func TestEndpointSource(t *testing.T) {
	assert := assert.New(t)

	proxy := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/bucket/key")
	})(t)

	endpointSource := httpmock.New(func(s *httpmock.Server) {
		s.ExpectGet("/endpoint").Return(strings.TrimPrefix(proxy.URL(), "http://"))
	})(t)

	pool := storage.NewSessionPool(&config.Storage{
		StorageEndpoint:    "storage.mock",
		EndpointSourceHost: endpointSource.URL() + "/endpoint",
		AccessKeyId:        "mock_access_key",
		SecretAccessKey:    "mock_secret_key",

		StorageRegion: "us-east-1",

		StorageConcurrency: 1,
	})

	sess, err := pool.GetSession(context.TODO())
	assert.NoError(err)
	_, err = sess.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})
	assert.NoError(err)
}
