#!/bin/bash
set -x

# Cat a non-existent object: server returns a permanent (404) error
# immediately instead of retrying. Client should exit non-zero with
# no payload output.
set +e
yp-client --config test/regress/conf/yproxy.yaml -l fatal cat 'missing_file' 2>/dev/null
rc=$?
set -e

if [ "$rc" -eq 0 ]; then
    echo "unexpected success"
    exit 1
fi

# Storage should still be empty
yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''
