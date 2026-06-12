#!/bin/bash
set -e

GO_DIR="/home/tony/.gobass_go"
GOPATH_DIR="/home/tony/.gobass_gopath"

echo "Downloading Go 1.24.2..."
mkdir -p "$GO_DIR"

if [ ! -f "$GO_DIR/go.tar.gz" ]; then
    curl -L -o "$GO_DIR/go.tar.gz" https://dl.google.com/go/go1.24.2.linux-amd64.tar.gz
else
    echo "Go tarball already downloaded, skipping download."
fi

echo "Extracting Go..."
# Clean the extraction directory first to avoid conflicts
rm -rf "$GO_DIR/go"
tar -C "$GO_DIR" -xzf "$GO_DIR/go.tar.gz"

# Set environment paths
export GOROOT="$GO_DIR/go"
export PATH="$GOROOT/bin:$PATH"
export GOPATH="$GOPATH_DIR"

echo "Go version:"
go version

if [ ! -f go.mod ]; then
    echo "Initializing Go module..."
    go mod init gobass
fi

echo "Getting MIDI package..."
go get gitlab.com/gomidi/midi/v2@latest

echo "Tidying module dependencies..."
go mod tidy

echo "Running tests..."
go test -v ./...

echo "Building project..."
go build -o gobass main.go

echo "Done! The 'gobass' executable has been built successfully."
