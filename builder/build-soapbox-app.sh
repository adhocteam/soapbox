#!/bin/sh

docker build -t ${APPLICATION_SLUG}:latest /build
docker save ${APPLICATION_SLUG}:latest | gzip > /build/${APPLICATION_SLUG}.tar.gz
