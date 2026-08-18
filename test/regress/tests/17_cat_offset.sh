#!/bin/bash
set -ex

echo 'offset test data here' | yp-client --config test/regress/conf/yproxy.yaml -l fatal put 'offset_file'
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
# Skip first 7 bytes ("offset "), read the rest
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat --offset 7 'offset_file'
