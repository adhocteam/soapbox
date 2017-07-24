#!/bin/bash

RELEASE=${RELEASE-latest}
if [[ $RELEASE =~ ^[a-fA-F0-9]{40}$ ]]
then
  # IF $RELEASE is a committish, use the first seven characters
  RELEASE=$(echo $RELEASE | awk '{print substr($0,0,7)}')
fi

docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  -v ${PATH_TO_CODE}:/build \
  -e APPLICATION_SLUG=${APPLICATION_SLUG} \
  -e RELEASE=${RELEASE} \
  -it soapbox

# Upload "latest" archive
echo 'aws s3 cp ${PATH_TO_CODE}/${APPLICATION_SLUG}-latest.tar.gz \
  s3://${S3_BUCKET-soapbox-app-images}/${APPLICATION_SLUG}/${APPLICATION_SLUG}-latest.tar.gz'

if [[ "$RELEASE" != "latest" ]]
then
  # Upload the archive with $RELEASE in filename
  echo 'aws s3 cp ${PATH_TO_CODE}/${APPLICATION_SLUG}-${RELEASE}.tar.gz \
    s3://${S3_BUCKET-soapbox-app-images}/${APPLICATION_SLUG}/${APPLICATION_SLUG}-${RELEASE}.tar.gz'
fi
