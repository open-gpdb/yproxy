#!/bin/bash
set -ex

echo 'plain delete data' | yp-client --config test/regress/conf/yproxy.yaml -l fatal put 'plain_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal delete --confirm 'plain_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
