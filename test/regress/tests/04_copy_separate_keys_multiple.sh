#!/bin/bash
set -ex

psql -h pg -U postgres -d test -c "INSERT INTO yezzey.yezzey_virtual_index VALUES ('/encrypted_file_old_key_1'), ('/encrypted_file_old_key_2')"

echo 'Sample encrypted data, old key' | yp-client --config test/regress/conf/yproxy_old.yaml -l fatal put --encrypt 'encrypted_file_old_key_1'
echo 'Sample encrypted data, old key' | yp-client --config test/regress/conf/yproxy_old.yaml -l fatal put --encrypt 'encrypted_file_old_key_2'

# Copy file encrypted with separate keys
yp-client copy --config test/regress/conf/yproxy.yaml --old-config test/regress/conf/yproxy_old.yaml --encrypt --decrypt --port 5432 --confirm --log-level error ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat --decrypt 'encrypted_file_old_key_1'
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat --decrypt 'encrypted_file_old_key_2'
