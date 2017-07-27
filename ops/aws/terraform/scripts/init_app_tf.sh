#!/bin/bash

set -e

STATE_BUCKET="${S3_BUCKET-soapbox-app-tf-state}"
REGION="${REGION-us-east-1}"
DYNAMO_TABLE="${DYNAMO_TABLE-soapbox-app-state-locking}"
PWD=$(pwd)

usage() {
    echo "Usage: $(basename $0) [-a application] [-e environment] [-t <network|deployment>]" 1>&2
    exit 1
}

while getopts ":a:e:t:" opt; do
    case "$opt" in
	a)
	    APP="$OPTARG"
	    ;;
	e)
	    ENV="$OPTARG"
	    ;;
        t)
            TYPE="$OPTARG"
            ;;
        *)
            usage
            ;;
    esac
done

if [ -z "$APP" ] || [ -z "$ENV" ] || [ -z "$TYPE" ]; then
    usage
fi

# Get a temp dir, copy the required tf files
APP_TMP=$(mktemp -d)
cp -R $PWD/../$TYPE/ $APP_TMP/

# Change to the temp dir
if [ "$TYPE" = "deployment" ]; then
  cd $APP_TMP/asg
else
  cd $APP_TMP
fi

# Populate backend.tfvar.sample with appropriate values
sed -i .bak "s/#APP#/$APP/" backend.tfvars.sample
sed -i .bak "s/#ENV#/$ENV/" backend.tfvars.sample
sed -i .bak "s/#STATE_BUCKET#/$STATE_BUCKET/" backend.tfvars.sample
sed -i .bak "s/#REGION#/$REGION/" backend.tfvars.sample
sed -i .bak "s/#DYNAMO_TABLE#/$DYNAMO_TABLE/" backend.tfvars.sample
rm -f *.bak

# Ceremonial promotion of .sample to bonafide .tfvars file
mv backend.tfvars.sample backend.tfvars

# Get any required terraform modules
terraform get

# Initialize the backend
terraform init -backend-config=backend.tfvars

# Use $APP_TMP for whatever operations you need to perform,
# cleaning up the temp dir when you're done
echo "temp: $APP_TMP"
