#!/bin/bash

function cleanup() {
    kill -USR2 $pid1
    kill -USR2 $pid2
    kill -ABRT $clientpid1
    kill -ABRT $clientpid2
    wait $pid1
    wait $pid2
}

devbin/yproxy --config examples/yproxy.yaml 2>&1 & pid1=$!
sleep 5s
if ! kill -0 $pid1 2>&1; then
    cleanup
    exit 1
fi
echo 'Instance 1 is running with pid' $pid1

devbin/yproxy --config examples/yproxy.yaml 2>&1 & pid2=$!
sleep 5s
if ! kill -0 $pid2 2>&1; then
    cleanup
    exit 1
fi
echo 'Instance 2 is running with pid' $pid2

devbin/yp-client --config examples/yproxy.yaml gool asdf 2>&1 & clientpid1=$!
if ! kill -0 $clientpid1 2>&1; then
    cleanup
    exit 1
fi
echo 'yp-client1 received connection with pid' $clientpid1

echo 'Killing Instance 1'
kill -USR2 $pid1
# old instance should be still running, serving active connections
if ! kill -0 $pid1 2>&1; then
    cleanup
    exit 1
fi
# client should still have a connection to old instance
if ! kill -0 $clientpid1 2>&1; then
    cleanup
    exit 1
fi

# new clients must be served
devbin/yp-client --config examples/yproxy.yaml gool asdf 2>&1 & clientpid2=$!
if ! kill -0 $clientpid2 2>&1; then
    cleanup
    exit 1
fi
echo 'yp-client2 received connection with pid' $clientpid2

echo 'Killing Instance 2'
kill -USR2 $pid2

cleanup

exit 0
