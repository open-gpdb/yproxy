#!/bin/bash
set -ex

# Force multipart upload with a tiny chunk size so 30 bytes split into multiple parts
printf 'multipart chunk test data here' | yp-client --config test/regress/conf/yproxy.yaml -l fatal put --multipart-chunk-size 8 'multipart_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat 'multipart_file'
