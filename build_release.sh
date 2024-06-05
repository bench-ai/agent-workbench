#!/bin/bash

# Package name and output directory
PACKAGE_NAME="agent"
OUTPUT_DIR="build"

# Supported OS and architecture combinations
OS_ARCH_COMBINATIONS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Build the package for each OS and architecture combination
for COMBO in "${OS_ARCH_COMBINATIONS[@]}"; do
    IFS="/" read -r OS ARCH <<< "$COMBO"
    OUTPUT_NAME="${PACKAGE_NAME}-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "Building for $OS/$ARCH..."
    GOOS=$OS GOARCH=$ARCH go build -o $OUTPUT_DIR/$OUTPUT_NAME

    if [ $? -ne 0 ]; then
        echo "Failed to build for $OS/$ARCH"
    else
        echo "Successfully built $OUTPUT_NAME"
    fi
done

OUTPUT_NAME="lambda-agent-linux-amd64"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $OUTPUT_DIR/$OUTPUT_NAME

echo "Build process completed."