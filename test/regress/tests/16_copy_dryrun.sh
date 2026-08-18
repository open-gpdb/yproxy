#!/bin/bash
set -ex

psql -h pg -U postgres -d test -c "INSERT INTO yezzey.yezzey_virtual_index VALUES ('/copy_dry_file')"

echo 'copy dry run data' | yp-client --config test/regress/conf/yproxy_old.yaml -l fatal put 'copy_dry_file'
yp-client --config test/regress/conf/yproxy_old.yaml -l fatal list ''

# Dry-run: no --confirm, nothing should be copied to destination
yp-client copy --config test/regress/conf/yproxy.yaml --old-config test/regress/conf/yproxy_old.yaml --port 5432 --log-level error 'copy_dry_file'

# Destination should be empty
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
# Source file should still be present
yp-client --config test/regress/conf/yproxy_old.yaml -l fatal list ''
