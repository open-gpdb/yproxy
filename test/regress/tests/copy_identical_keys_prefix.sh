#!/bin/bash
set -e

echo 'Sample encrypted data, identical key' | yp-client --config test/regress/conf/yproxy_second_bucket.yaml -l fatal put --encrypt 'encrypted_file_identical_key'
yp-client --config test/regress/conf/yproxy_second_bucket.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_second_bucket.yaml -l fatal cat --decrypt 'encrypted_file_identical_key'

psql -h pg -U postgres -d test -c "INSERT INTO yezzey.yezzey_virtual_index VALUES ('/encrypted_file_identical_key')"

yp-client copy --config test/regress/conf/yproxy_prefix.yaml --old-config test/regress/conf/yproxy_second_bucket.yaml --encrypt --decrypt --port 5432 --confirm --log-level error 'encrypted_file_identical_key'
yp-client --config test/regress/conf/yproxy_prefix.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_prefix.yaml -l fatal cat --decrypt 'encrypted_file_identical_key'
