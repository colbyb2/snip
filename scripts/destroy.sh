#!/bin/bash

set -e

echo "Destroying Terraform Environment"

echo "Switching to Terraform Directory"

cd ../terraform

terraform destroy

echo "Terraform Environment Destroyed"
