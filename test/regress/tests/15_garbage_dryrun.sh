#!/bin/bash
set -ex

echo 'garbage dry data' | yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal put 'garbage_dry_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
# Dry-run: no --confirm, file must remain in place
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal delete --garbage --port 5432 --segnum 0 'garbage_dry_file'
yp-client --config test/regress/conf/yproxy_vacuum.yaml -l fatal list ''
