#!/bin/sh
set -ex
export GO111MODULE=on
go get -v -tags netgo -ldflags "-extldflags \"-static\"" ./cmd/...
