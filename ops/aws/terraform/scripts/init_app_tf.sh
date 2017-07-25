#!/bin/bash

set -e

APP="${APP_NAME-soapbox-app-name}"
ENV="${ENV-test}"
STATE_BUCKET="${S3_BUCKET-soapbox-app-state}"
REGION="${REGION-us-east-1}"
DYNAMO_TABLE="${DYNAMO_TABLE-soapbox-app-state-locking}"
PWD=$(pwd)

usage() { echo "Usage: $0 [-t <network|deployment>]" 1>&2; exit 1; }

while getopts ":t:" opt; do
    case "${opt}" in
        t)
            TYPE="${OPTARG}"
            ;;
        *)
            usage
            ;;
    esac
done

if [[ "$TYPE" == "" ]]
then
  usage;
fi

# Get a temp dir, copy the required tf files
APP_TMP=$(mktemp -d)
cp -R ${PWD}/../${TYPE}/ $APP_TMP/

# Change to the temp dir
if [[ "$TYPE" == "deployment" ]]
then
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
rm -Rf  *.bak

# Ceremonial promotion of .sample to bonafide .tfvars file
mv backend.tfvars.sample backend.tfvars

# Get any required terraform modules
terraform get

# Initialize the backed
terraform init -backend-config=backend.tfvars

# Use $APP_TMP for whatever operations you need to perform,
# cleaning up the temp dir when you're done
echo "temp: $APP_TMP"
