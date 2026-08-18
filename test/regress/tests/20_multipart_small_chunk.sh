#!/bin/bash
set -ex


# Multipart should be at least 5242880 bytes, so we generate 3 full 
# part and one small tail
set +x
value=$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c $((5242880 * 3 + 67)))

# Force multipart upload with a tiny chunk size so 30 bytes split into multiple parts
printf $value | yp-client --config test/regress/conf/yproxy.yaml -l fatal put --multipart-chunk-size 5242880 'multipart_file'
value_cat=$(yp-client --config test/regress/conf/yproxy.yaml -l fatal cat 'multipart_file')

set -x

yp-client --config test/regress/conf/yproxy.yaml -l fatal list ''

set +x
[ $value -eq $value_cat ] || exit 1
set -x

echo ok
