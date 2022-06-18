FROM golang:1.18.3-alpine as build
RUN apk add --no-cache git
ADD . /go/src/github.com/neubot/dash
WORKDIR /go/src/github.com/neubot/dash
RUN ./build.sh

FROM gcr.io/distroless/static@sha256:2ad95019a0cbf07e0f917134f97dd859aaccc09258eb94edcb91674b3c1f448f
COPY --from=build /go/bin/dash-server /
EXPOSE 80/tcp 443/tcp
WORKDIR /
ENTRYPOINT ["/dash-server"]
