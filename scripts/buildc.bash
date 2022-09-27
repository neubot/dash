#!/bin/bash
set -euxo pipefail
docker build -t neubot/dash .
docker tag neubot/dash neubot/dash:$(git describe --tags --dirty)-$(date -u +%Y%m%d%H%M%S)
