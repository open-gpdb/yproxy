#!/bin/bash
set -ex

echo 'trash data' | yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal put 'trash/garbage_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal deleteTrash 'trash' --confirm
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
