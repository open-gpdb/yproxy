#!/bin/bash
set -ex

echo 'garbage data' | yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal put 'some_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal delete --garbage --confirm --port 5432 --segnum 0 'some_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''

s3cmd --access_key some_key --secret_key some_key --host minio:9000 --host-bucket '' --no-ssl info s3://gpyezzey/trash/segments_005/seg0/basebackups_005/yezzey/some_file | grep Storage
