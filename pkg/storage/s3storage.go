package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/yezzey-gp/aws-sdk-go/aws"
	"github.com/yezzey-gp/aws-sdk-go/service/s3"
	"github.com/yezzey-gp/aws-sdk-go/service/s3/s3manager"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/message"
	"github.com/yezzey-gp/yproxy/pkg/object"
	"github.com/yezzey-gp/yproxy/pkg/settings"
	"github.com/yezzey-gp/yproxy/pkg/tablespace"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"
)

type S3StorageInteractor struct {
	StorageInteractor

	pool SessionPool

	cnf *config.Storage

	bucketMap        map[string]string
	multipartUploads sync.Map
}

func (s *S3StorageInteractor) CatFileFromStorage(name string, offset int64, setts []settings.StorageSettings) (io.ReadCloser, error) {
	// XXX: fix this
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return nil, err
	}

	objectPath := strings.TrimLeft(path.Join(s.cnf.StoragePrefix, name), "/")

	tableSpace := ResolveStorageSetting(setts, message.TableSpaceSetting, tablespace.DefaultTableSpace)

	bucket, ok := s.bucketMap[tableSpace]
	if !ok {
		err := fmt.Errorf("failed to match tablespace %s to s3 bucket", tableSpace)
		ylogger.Zero.Err(err)
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    aws.String(objectPath),
		Range:  aws.String(fmt.Sprintf("bytes=%d-", offset)),
	}

	ylogger.Zero.Debug().Str("key", objectPath).Int64("offset", offset).Str("bucket",
		s.cnf.StorageBucket).Msg("requesting external storage")

	object, err := sess.GetObject(input)
	return object.Body, err
}

func (s *S3StorageInteractor) PutFileToDest(name string, r io.Reader, settings []settings.StorageSettings) error {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return err
	}

	objectPath := strings.TrimLeft(path.Join(s.cnf.StoragePrefix, name), "/")

	storageClass := ResolveStorageSetting(settings, message.StorageClassSetting, "STANDARD")
	tableSpace := ResolveStorageSetting(settings, message.TableSpaceSetting, tablespace.DefaultTableSpace)
	multipartChunksizeStr := ResolveStorageSetting(settings, message.MultipartChunksize, "16777216")
	multipartChunksize, err := strconv.ParseInt(multipartChunksizeStr, 10, 64)
	if err != nil {
		return err
	}
	multipartUpload, err := strconv.ParseBool(ResolveStorageSetting(settings, message.MultipartUpload, "1"))
	if err != nil {
		return err
	}

	up := s3manager.NewUploaderWithClient(sess, func(uploader *s3manager.Uploader) {
		uploader.PartSize = int64(multipartChunksize)
		uploader.Concurrency = 1
	})

	bucket, ok := s.bucketMap[tableSpace]
	if !ok {
		err := fmt.Errorf("failed to match tablespace %s to s3 bucket", tableSpace)
		ylogger.Zero.Err(err)
		return err
	}

	if multipartUpload {
		s.multipartUploads.Store(objectPath, true)
		_, err = up.Upload(
			&s3manager.UploadInput{
				Bucket:       aws.String(bucket),
				Key:          aws.String(objectPath),
				Body:         r,
				StorageClass: aws.String(storageClass),
			},
		)
		s.multipartUploads.Delete(objectPath)
	} else {
		var body []byte
		body, err = io.ReadAll(r)
		if err != nil {
			return err
		}
		_, err = sess.PutObject(&s3.PutObjectInput{
			Bucket:       aws.String(bucket),
			Key:          aws.String(objectPath),
			Body:         bytes.NewReader(body),
			StorageClass: aws.String(storageClass),
		})
	}

	return err
}

func (s *S3StorageInteractor) PatchFile(name string, r io.ReadSeeker, startOffset int64) error {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return nil
	}

	objectPath := strings.TrimLeft(path.Join(s.cnf.StoragePrefix, name), "/")

	input := &s3.PatchObjectInput{
		Bucket:       &s.cnf.StorageBucket,
		Key:          aws.String(objectPath),
		Body:         r,
		ContentRange: aws.String(fmt.Sprintf("bytes %d-18446744073709551615", startOffset)),
	}

	_, err = sess.PatchObject(input)

	ylogger.Zero.Debug().Str("key", objectPath).Str("bucket",
		s.cnf.StorageBucket).Msg("modifying file in external storage")

	return err
}

func (s *S3StorageInteractor) ListPath(prefix string) ([]*object.ObjectInfo, error) {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return nil, err
	}

	var continuationToken *string
	prefix = strings.TrimLeft(path.Join(s.cnf.StoragePrefix, prefix), "/")
	metas := make([]*object.ObjectInfo, 0)

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            &s.cnf.StorageBucket,
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		out, err := sess.ListObjectsV2(input)
		if err != nil {
			ylogger.Zero.Debug().Err(err).Msg("failed to list prefix")
			return nil, err
		}

		for _, obj := range out.Contents {
			path := *obj.Key

			cPath, ok := strings.CutPrefix(path, s.cnf.StoragePrefix)
			if !ok {
				ylogger.Zero.Debug().Str("path", path).Msg("skipping file")
				continue
			}
			ylogger.Zero.Debug().Str("path", path).Msg("appending file")
			metas = append(metas, &object.ObjectInfo{
				Path: "/" + cPath,
				Size: *obj.Size,
			})
		}

		if !*out.IsTruncated {
			break
		}

		continuationToken = out.NextContinuationToken
	}
	return metas, nil
}

func (s *S3StorageInteractor) DeleteObject(key string) error {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return err
	}
	ylogger.Zero.Debug().Msg("aquired session")

	if !strings.HasPrefix(key, s.cnf.StoragePrefix) {
		key = path.Join(s.cnf.StoragePrefix, key)
	}
	key = strings.TrimLeft(key, "/")

	input2 := s3.DeleteObjectInput{
		Bucket: &s.cnf.StorageBucket,
		Key:    aws.String(key),
	}

	_, err = sess.DeleteObject(&input2)
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to delete old object")
		return err
	}
	ylogger.Zero.Debug().Str("path", key).Msg("deleted object")
	return nil
}

func (s *S3StorageInteractor) SScopyObject(from string, to string) error {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to acquire s3 session")
		return err
	}
	ylogger.Zero.Debug().Msg("aquired session for server-side copy")

	if !strings.HasPrefix(from, s.cnf.StoragePrefix) {
		from = path.Join(s.cnf.StoragePrefix, from)
	}
	from = strings.TrimLeft(from, "/")

	if !strings.HasPrefix(to, s.cnf.StoragePrefix) {
		to = path.Join(s.cnf.StoragePrefix, to)
	}
	to = strings.TrimLeft(to, "/")

	inp := s3.CopyObjectInput{
		Bucket:     &s.cnf.StorageBucket,
		CopySource: aws.String(path.Join(s.cnf.StorageBucket, from)),
		Key:        aws.String(to),
	}

	_, err = sess.CopyObject(&inp)
	if err != nil {
		ylogger.Zero.Err(err).Msg("failed to copy object")
		return err
	}
	ylogger.Zero.Debug().Str("path-from", from).Str("path-to", to).Msg("copied object")

	return nil
}

func (s *S3StorageInteractor) MoveObject(from string, to string) error {
	if err := s.SScopyObject(from, to); err != nil {
		return err
	}
	return s.DeleteObject(from)
}

func (s *S3StorageInteractor) AbortMultipartUpload(key, uploadId string) error {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		return err
	}

	_, err = sess.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.cnf.StorageBucket),
		UploadId: aws.String(uploadId),
		Key:      aws.String(key),
	})
	return err
}

func (s *S3StorageInteractor) ListFailedMultipartUploads() (map[string]string, error) {
	sess, err := s.pool.GetSession(context.TODO())
	if err != nil {
		return nil, err
	}

	uploads := make([]*s3.MultipartUpload, 0)
	var keyMarker *string
	for {
		out, err := sess.ListMultipartUploads(&s3.ListMultipartUploadsInput{
			Bucket:    aws.String(s.cnf.StorageBucket),
			KeyMarker: keyMarker,
		})
		if err != nil {
			return nil, err
		}

		uploads = append(uploads, out.Uploads...)

		if !*out.IsTruncated {
			break
		}

		keyMarker = out.NextKeyMarker
	}

	out := make(map[string]string)
	for _, upload := range uploads {
		if _, ok := s.multipartUploads.Load(*upload.Key); !ok {
			out[*upload.Key] = *upload.UploadId
		}
	}
	return out, nil
}
