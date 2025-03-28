#!/bin/bash
set -ex

yp-client --config test/regress/yproxy.yaml -l fatal list ''
echo 'Sample data' | yp-client --config test/regress/yproxy.yaml -l fatal put 'some_file'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml  -l fatal cat 'some_file'

echo 'Sample encrypted data' | yp-client --config test/regress/yproxy.yaml -l fatal put --encrypt 'encrypted_file'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml  -l fatal cat --decrypt 'encrypted_file'

# Put file in the second bucket, encryption with separate key
echo 'Sample encrypted data, old key' | yp-client --config test/regress/yproxy_old.yaml -l fatal put --encrypt 'encrypted_file_old_key'
yp-client --config test/regress/yproxy_old.yaml -l fatal list ''
yp-client --config test/regress/yproxy_old.yaml  -l fatal cat --decrypt 'encrypted_file_old_key'

# Copy file encrypted with separate keys
yp-client copy --config test/regress/yproxy.yaml --old-config test/regress/yproxy_old.yaml --encrypt --decrypt --port 5432 --confirm --log-level error 'encrypted_file_old_key'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml  -l fatal cat --decrypt 'encrypted_file_old_key'
