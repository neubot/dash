# Neubot DASH Go client and server

[![GoDoc](https://godoc.org/github.com/neubot/dash?status.svg)](https://godoc.org/github.com/neubot/dash) [![Build Status](https://travis-ci.org/neubot/dash.svg?branch=master)](https://travis-ci.org/neubot/dash) [![Coverage Status](https://coveralls.io/repos/github/neubot/dash/badge.svg?branch=master)](https://coveralls.io/github/neubot/dash?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/neubot/dash)](https://goreportcard.com/report/github.com/neubot/dash)

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

```bash
docker run --network=host                    \
           --volume `pwd`/datadir:/datadir   \
           --volume `pwd`/cache:/root/.cache \
           --read-only                       \
           --cap-drop=all                    \
           --cap-add=net_bind_service        \
           neubot/dash                       \
           -datadir /datadir
```

### Release

```bash
docker push neubot/dash
```
