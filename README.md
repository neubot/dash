# Neubot DASH Go client and server

[![GoDoc](https://godoc.org/github.com/neubot/dash?status.svg)](https://godoc.org/github.com/neubot/dash) ![Golang Status](https://github.com/neubot/dash/workflows/golang/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/neubot/dash/badge.svg?branch=master)](https://coveralls.io/github/neubot/dash?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/neubot/dash)](https://goreportcard.com/report/github.com/neubot/dash)

This repository contains an implementation of Neubot's DASH experiment
client and server, both written in Go.

## Server

The server is meant to be deployed at Measurement Lab. For this reason the
release procedure for the server, described below, uses Docker. Images will
be available as [neubot/dash](https://hub.docker.com/r/neubot/dash).

### Build

```bash
docker build -t neubot/dash .
docker tag neubot/dash neubot/dash:`git describe --tags --dirty`-`date -u +%Y%m%d%H%M%S`
```

### Test locally

The following command should work on a Linux system:

```bash
rm -f ./certs/cert.pem ./certs/key.pem &&    \
./mkcerts.bash &&                            \
sudo chown root:root ./certs/*.pem &&        \
docker run --network=bridge                  \
           --publish=80:8888                 \
           --publish=443:4444                \
           --publish=9990:9999               \
           --volume `pwd`/certs:/certs:ro    \
           --volume `pwd`/datadir:/datadir   \
           --read-only                       \
           --cap-drop=all                    \
           neubot/dash                       \
           -datadir /datadir                 \
           -http-listen-address :8888        \
           -https-listen-address :4444       \
           -prometheusx.listen-address :9999 \
           -tls-cert /certs/cert.pem         \
           -tls-key /certs/key.pem
```

This command will run `dash-server` in a container as the root user, with
no capabilities, limiting access to the file system and exposing all the
relevant ports: 80 for HTTP based tests, 443 for HTTPS tests, and 9990 to
access prometheus metrics.

### Release

```bash
docker push neubot/dash
```

## Client

Build using:

```bash
go build -v ./cmd/dash-client
```

Make sure you read [PRIVACY.md](PRIVACY.md) before running. The command
will anyway refuse to run unless you acknowledge the privacy policy by
passing the `-y` command line flag.
