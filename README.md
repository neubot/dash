# Neubot DASH Go client and server

[![GoDoc](https://godoc.org/github.com/neubot/dash?status.svg)](https://godoc.org/github.com/neubot/dash) ![Golang Status](https://github.com/neubot/dash/workflows/golang/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/neubot/dash/badge.svg?branch=master)](https://coveralls.io/github/neubot/dash?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/neubot/dash)](https://goreportcard.com/report/github.com/neubot/dash)

This repository contains an implementation of Neubot's DASH experiment
client and server, both written in Go.

## Server

The server is meant to be deployed at Measurement Lab. For this reason the
release procedure for the server, described below, uses Docker. Images will
be available as [neubot/dash](https://hub.docker.com/r/neubot/dash).

### Build Docker Container

```bash
make buildcontainer
```

### Test Docker Container locally

The following command should work on a Linux system:

```bash
make runcontainer
```

This command will run `dash-server` in a container as the root user, with
no capabilities, limiting access to the file system and exposing all the
relevant ports: 80 for HTTP based tests, 443 for HTTPS tests, and 9990 to
access prometheus metrics.

### Release

To push the container at DockerHub, run:

```bash
docker push neubot/dash
```

The procedure to update the version that runs on M-Lab is the following:

1. open a pull request at m-lab/dash so they know they need to sync from this repo
2. ask the m-lab staff to pull the tagged version
3. open a pull request [for neubot.jsonnet at m-lab/k8s-support](https://github.com/m-lab/k8s-support/blob/master/k8s/daemonsets/experiments/neubot.jsonnet#L17)

At this point, `neubot/dash` is deployed in staging. To test, use

```
https://locate-dot-mlab-staging.appspot.com/v2/nearest/neubot/dash
```

as the locate URL instead of the canonical one.

## Client

Build using:

```bash
go build -v ./cmd/dash-client
```

Make sure you read [PRIVACY.md](PRIVACY.md) before running. The command
will anyway refuse to run unless you acknowledge the privacy policy by
passing the `-y` command line flag.
