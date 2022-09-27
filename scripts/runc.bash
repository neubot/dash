#!/bin/bash
set -euxo pipefail
rm -f ./certs/cert.pem ./certs/key.pem
./scripts/mkcerts.bash
sudo chown root:root ./certs/*.pem
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
