#!/bin/bash
set -e

echo 'Sample encrypted data' | yp-client --config test/regress/conf/yproxy.yaml -l fatal put --encrypt 'encrypted_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat --decrypt 'encrypted_file'
