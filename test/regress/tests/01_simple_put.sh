#!/bin/bash
set -e

echo 'Sample data' | yp-client --config test/regress/conf/yproxy.yaml -l fatal put 'some_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat 'some_file'
