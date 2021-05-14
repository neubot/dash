FROM golang:1.15.0-alpine3.12 as build
RUN apk add --no-cache git
ADD . /go/src/github.com/neubot/dash
WORKDIR /go/src/github.com/neubot/dash
RUN ./build.sh

FROM gcr.io/distroless/static@sha256:3cd546c0b3ddcc8cae673ed078b59067e22dea676c66bd5ca8d302aef2c6e845
COPY --from=build /go/bin/dash-server /
EXPOSE 80/tcp 443/tcp
WORKDIR /
ENTRYPOINT ["/dash-server"]
