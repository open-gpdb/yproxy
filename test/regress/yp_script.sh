#!/bin/bash
set -ex

yp-client --config test/regress/yproxy.yaml -l fatal list ''
echo 'Sample data' | yp-client --config test/regress/yproxy.yaml -l fatal put 'some_file'
yp-client --config test/regress/yproxy.yaml -l fatal list ''
yp-client --config test/regress/yproxy.yaml  -l fatal cat 'some_file'
