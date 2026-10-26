package storage_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yezzey-gp/aws-sdk-go/aws"
	"github.com/yezzey-gp/aws-sdk-go/aws/client"
	"github.com/yezzey-gp/aws-sdk-go/service/s3"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/storage"
)

type retryTransportFunc func(*http.Request) (*http.Response, error)

func (f retryTransportFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestS3UploadRetryWithoutExpect(t *testing.T) {
	for _, operation := range []string{"PutObject", "UploadPart"} {
		t.Run(operation, func(t *testing.T) {
			pool := storage.NewSessionPool(&config.Storage{
				StorageEndpoint:    "http://storage.example",
				StorageRegion:      "us-east-1",
				StorageConcurrency: 1,
			}, "")
			svc, err := pool.GetSession(context.Background(), &config.StorageCredentials{
				AccessKeyId: "test-access", SecretAccessKey: "test-secret",
			})
			if err != nil {
				t.Fatal(err)
			}
			payload := bytes.Repeat([]byte("x"), 5*1024*1024)
			attempts := 0
			svc.Retryer = client.DefaultRetryer{
				NumMaxRetries: 1, MinRetryDelay: time.Nanosecond, MaxRetryDelay: time.Nanosecond,
			}
			svc.Config.HTTPClient = &http.Client{Transport: retryTransportFunc(func(r *http.Request) (*http.Response, error) {
				attempts++
				if got := r.Header.Get("Expect"); got != "" {
					t.Errorf("attempt %d: unexpected Expect header %q", attempts, got)
				}
				authorization := r.Header.Get("Authorization")
				if !strings.Contains(authorization, "SignedHeaders=") {
					t.Errorf("attempt %d: missing signature", attempts)
				}
				if strings.Contains(authorization, "expect") {
					t.Errorf("attempt %d: Expect included in signature: %s", attempts, authorization)
				}
				body, err := io.ReadAll(r.Body)
				_ = r.Body.Close()
				if err != nil || !bytes.Equal(body, payload) {
					t.Errorf("attempt %d: payload not replayed correctly: %v", attempts, err)
				}
				status, response := http.StatusOK, ""
				if attempts == 1 {
					status, response = http.StatusInternalServerError, "<Error><Code>InternalError</Code></Error>"
				}
				return &http.Response{
					StatusCode: status, Header: http.Header{"Etag": []string{"test-etag"}},
					Body: io.NopCloser(strings.NewReader(response)), Request: r,
				}, nil
			})}
			if operation == "PutObject" {
				_, err = svc.PutObject(&s3.PutObjectInput{
					Bucket: aws.String("bucket"), Key: aws.String("key"), Body: bytes.NewReader(payload),
				})
			} else {
				_, err = svc.UploadPart(&s3.UploadPartInput{
					Bucket: aws.String("bucket"), Key: aws.String("key"), Body: bytes.NewReader(payload),
					UploadId: aws.String("test-upload"), PartNumber: aws.Int64(1),
				})
			}
			if err != nil {
				t.Fatal(err)
			}
			if attempts != 2 {
				t.Fatalf("expected initial request and one retry, got %d attempts", attempts)
			}
		})
	}
}
