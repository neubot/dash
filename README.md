# Neubot DASH Go client and server

This repository contains an implementation of Neubot's DASH experiment
client and server, both written in Go.

## Server

The server is meant to be deployed at Measurement Lab. For this reason the
release procedure for the server, described below, uses Docker.

### Build

```bash
docker build -t neubot/dash .
```

### Test locally

```bash
install -d datadir && docker run --network=bridge                \
           			 --publish 80:8880               \
           			 --volume `pwd`/datadir:/datadir \
           			 --read-only                     \
           			 --user `id -u`:`id -g`          \
           			 --cap-drop=all                  \
           			 neubot/dash                     \
           			 -datadir /datadir               \
           			 -endpoint :8880
```

### Release

```bash
docker tag neubot/dash neubot/dash:`date -u +%Y%m%d%H%M%S`
docker push neubot/dash
```
