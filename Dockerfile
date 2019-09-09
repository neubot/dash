FROM golang:1.13.0-alpine3.10 as build
RUN apk add --no-cache git
ADD . /go/src/github.com/neubot/dash
WORKDIR /go/src/github.com/neubot/dash
RUN ./build.sh

FROM gcr.io/distroless/static@sha256:9b60270ec0991bc4f14bda475e8cae75594d8197d0ae58576ace84694aa75d7a
COPY --from=build /go/bin/dash-server /
EXPOSE 80/tcp 443/tcp
WORKDIR /
ENTRYPOINT ["/dash-server"]
