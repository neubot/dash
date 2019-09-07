FROM golang:alpine as build
RUN apk add --no-cache git
ADD . /go/src/github.com/neubot/dash
WORKDIR /go/src/github.com/neubot/dash
RUN ./build.sh

FROM gcr.io/distroless/static
COPY --from=build /go/bin/dash-server /
WORKDIR /
ENTRYPOINT ["/dash-server"]
