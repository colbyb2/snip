#!/bin/bash

set -e

echo "Generating Terraform Environment"

echo "Building Lambda Zip"

./build-lambda.sh

echo "Switching to Terraform Directory"

cd ../terraform

terraform apply -auto-approve

echo "Terraform Environment Generated"
