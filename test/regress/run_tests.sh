#!/bin/bash
set -ex

yproxy --config test/regress/conf/yproxy.yaml --test-mode &
yproxy --config test/regress/conf/yproxy_prexix.yaml --test-mode &
yproxy --config test/regress/conf/yproxy_old.yaml --test-mode &
yproxy --config test/regress/conf/yproxy_second_bucket.yaml --test-mode &

# Wait for yproxy to become available
i=0
while (! [ -S /tmp/yproxy.sock ] || ! [ -S /tmp/yproxy_old.sock ] || ! [ -S /tmp/yproxy_identical.sock ] || ! [ -S /tmp/yproxy_prefix.sock ] ) && [ $i -lt 20 ]; do sleep 1; i=$(($i+1)) ; done

for test in $(ls test/regress/tests | awk '{print(substr($1, 1, length($1)-3))}' )
    do ./test/regress/tests/${test}.sh > output.txt || true
    diff test/regress/expected/${test}.txt output.txt
    for file in $(s3cmd --access_key some_key --secret_key some_key --host minio:9000 --host-bucket "" --no-ssl la -r | awk '{print $4}' )
        do s3cmd --access_key some_key --secret_key some_key --host minio:9000 --host-bucket "" --no-ssl rm $file 
    done
    [ $(s3cmd --access_key some_key --secret_key some_key --host minio:9000 --host-bucket "" --no-ssl la -r | grep "\S" | wc -l ) -eq 0 ] || ( echo s3 not empty; exit 2 )
    psql -h pg -U postgres -d test -c "DELETE FROM yezzey.yezzey_virtual_index"
done
