#!/bin/bash
set -ex

echo 'garbage data' | yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal put 'some_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal delete --garbage --confirm --crazy-drop --port 5432 --segnum 0 'some_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
