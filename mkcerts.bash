#!/bin/bash
set -euxo pipefail
install -d certs
openssl genrsa -out certs/key.pem
openssl req -new -x509 -key key.pem -out certs/cert.pem -days 2 \
  -subj "/C=XX/ST=State/L=Locality/O=Org/OU=Unit/CN=localhost/emailAddress=test@email.address"