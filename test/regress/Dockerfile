FROM golang:latest

WORKDIR /root
COPY . /root

RUN mkdir /data

ENV PGHOST=pg
ENV PGUSER=postgres

RUN apt update && apt install -y postgresql-client s3cmd

RUN go build -pgo=auto -o /usr/bin/yproxy $(LDFLAGS) ./cmd/yproxy && \
    go build -o /usr/bin/yp-client ./cmd/client

ENTRYPOINT [ "test/regress/run_tests.sh" ]
