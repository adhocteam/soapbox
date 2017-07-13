#!/bin/bash

docker run -v /var/run/docker.sock:/var/run/docker.sock -v ${PATH_TO_CODE}:/build -e APPLICATION_SLUG=${APPLICATION_SLUG} -it soapbox
aws s3 cp ${PATH_TO_CODE}/${APPLICATION_SLUG}.tar.gz s3://soapbox-app-images/${APPLICATION_SLUG}/${APPLICATION_SLUG}.tar.gz