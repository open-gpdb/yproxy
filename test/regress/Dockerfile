FROM golang:latest

WORKDIR /root
COPY . /root

RUN mkdir /data

RUN go build -pgo=auto -o /usr/bin/yproxy $(LDFLAGS) ./cmd/yproxy && \
    go build -o /usr/bin/yp-client ./cmd/client

ENTRYPOINT [ "test/regress/run_tests.sh" ]
