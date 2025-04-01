#!/bin/bash
set -ex

yp-client --config test/regress/yproxy.yaml -l fatal list ''
echo 'Sample data' | yp-client --config test/regress/yproxy.yaml -l fatal put 'some_file'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml -l fatal cat 'some_file'

echo 'Sample encrypted data' | yp-client --config test/regress/yproxy.yaml -l fatal put --encrypt 'encrypted_file'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml -l fatal cat --decrypt 'encrypted_file'

# Put file in the second bucket, encryption with separate key
echo 'Sample encrypted data, old key' | yp-client --config test/regress/yproxy_old.yaml -l fatal put --encrypt 'encrypted_file_old_key'
yp-client --config test/regress/yproxy_old.yaml -l fatal list ''
yp-client --config test/regress/yproxy_old.yaml -l fatal cat --decrypt 'encrypted_file_old_key'

# Copy file encrypted with separate keys
yp-client copy --config test/regress/yproxy.yaml --old-config test/regress/yproxy_old.yaml --encrypt --decrypt --port 5432 --confirm --log-level error 'encrypted_file_old_key'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml -l fatal cat --decrypt 'encrypted_file_old_key'

# Test server-side copy

# Put file in the second bucket, encryption with the same key
echo 'Sample encrypted data, identical key' | yp-client --config test/regress/yproxy_second_bucket.yaml -l fatal put --encrypt 'encrypted_file_identical_key'
yp-client --config test/regress/yproxy_second_bucket.yaml -l fatal list ''
yp-client --config test/regress/yproxy_second_bucket.yaml -l fatal cat --decrypt 'encrypted_file_identical_key'

# Set up files to copy
psql -h pg -U postgres -d test -c "DELETE FROM yezzey.yezzey_virtual_index; INSERT INTO yezzey.yezzey_virtual_index VALUES ('/encrypted_file_identical_key')"

# Copy file encrypted with identical keys
yp-client copy --config test/regress/yproxy.yaml --old-config test/regress/yproxy_second_bucket.yaml --encrypt --decrypt --port 5432 --confirm --log-level error 'encrypted_file_identical_key'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml -l fatal cat --decrypt 'encrypted_file_identical_key'
