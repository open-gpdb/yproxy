package storage_test

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/yezzey-gp/aws-sdk-go/aws"
	"github.com/yezzey-gp/aws-sdk-go/service/s3"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/storage"
)

func TestEndpointSourceHTTP(t *testing.T) {
	assert := assert.New(t)

	httpmock.Activate()

	httpmock.RegisterResponder("GET", "http://endpoint_source/get_proxy",
		httpmock.NewStringResponder(200, "storage.proxy"))

	httpmock.RegisterResponder("GET", "http://storage.proxy/bucket/key",
		httpmock.NewStringResponder(200, ""))

	pool := storage.NewSessionPool(&config.Storage{
		StorageEndpoint:      "storage.mock",
		EndpointSourceHost:   "http://endpoint_source/get_proxy",
		EndpointSourceScheme: "http",
		AccessKeyId:          "mock_access_key",
		SecretAccessKey:      "mock_secret_key",

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

	info := httpmock.GetCallCountInfo()
	assert.Equal(1, info["GET http://endpoint_source/get_proxy"])
	assert.Equal(1, info["GET http://storage.proxy/bucket/key"])
}
