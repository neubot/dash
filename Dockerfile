FROM golang:1.24 as build
ADD . /go/src/github.com/neubot/dash
WORKDIR /go/src/github.com/neubot/dash
RUN CGO_ENABLED=0 go build -v -tags netgo -ldflags "-s -w -extldflags \"-static\"" ./cmd/dash-server

FROM gcr.io/distroless/static@sha256:2ad95019a0cbf07e0f917134f97dd859aaccc09258eb94edcb91674b3c1f448f
COPY --from=build /go/src/github.com/neubot/dash/dash-server /
EXPOSE 80/tcp 443/tcp
WORKDIR /
ENTRYPOINT ["/dash-server"]
