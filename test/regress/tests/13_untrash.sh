#!/bin/bash
set -ex

echo 'untrash data' | yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal put 'untrash_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal delete --garbage --confirm --port 5432 --segnum 0 'untrash_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''

# Trash path after garbage move: trash/segments_005/seg0/basebackups_005/yezzey/untrash_file
# Untrash restores to: segments_005/seg0/basebackups_005/yezzey/untrash_file
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal untrash 'trash/segments_005/seg0/basebackups_005/yezzey/untrash_file' --confirm --segnum 0
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal cat 'segments_005/seg0/basebackups_005/yezzey/untrash_file'
