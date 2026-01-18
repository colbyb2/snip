#!/bin/bash
set -e

echo "Building Lambda function..."

# Create build directory
mkdir -p build

# Build for Linux ARM64 (Lambda Graviton)
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o build/bootstrap cmd/lambda/main.go cmd/lambda/handlers.go cmd/lambda/dynamo.go

# Create zip file
cd build
zip -j lambda.zip bootstrap
cd ..

echo "Build complete: build/lambda.zip"