#!/bin/bash

echo "Running AWS CDK Deployer ..."

if [[ -z "${AWS_ACCESS_KEY_ID}" ]]; then
    echo "set AWS_ACCESS_KEY_ID env var"
    exit 1
fi

if [[ -z "${AWS_SECRET_ACCESS_KEY}" ]]; then
    echo "set AWS_SECRET_ACCESS_KEY env var"
    exit 1
fi

if [[ -z "${AWS_DEFAULT_REGION}" ]]; then
    echo "set AWS_DEFAULT_REGION env var"
    exit 1
fi

cdk bootstrap
cdk deploy LamassuCdkStack --outputs-file outputs.json --require-approval never
cat outputs.json