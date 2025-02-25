#!/bin/bash
set -ex

yproxy --config test/regress/yproxy.yaml &

i=0
while (! [ -S /tmp/yproxy.sock ]) && [ $i -lt 20 ]; do sleep 1; i=$(($i+1)) ; done

./test/regress/yp_script.sh > output.txt || true

diff test/regress/expected.txt output.txt
