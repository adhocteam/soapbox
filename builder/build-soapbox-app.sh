#!/bin/sh

docker build -t ${APPLICATION_SLUG}:latest /build
docker build -t ${APPLICATION_SLUG}:${RELEASE} /build

docker save ${APPLICATION_SLUG}:latest | gzip > /build/${APPLICATION_SLUG}-latest.tar.gz
if [[ "$RELEASE" != "latest" ]]
then
  docker save ${APPLICATION_SLUG}:${RELEASE} | gzip > /build/${APPLICATION_SLUG}-$RELEASE.tar.gz
fi
