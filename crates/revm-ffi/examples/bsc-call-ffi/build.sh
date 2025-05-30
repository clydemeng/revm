#!/bin/bash

set -e

echo "🔧 Building BSC Call FFI Example"

# Navigate to workspace root
cd ../../../../

# Build REVM FFI library in release mode
echo "📦 Building REVM FFI library..."
cargo build --release -p revm-ffi

# Navigate back to example directory
cd crates/revm-ffi/examples/bsc-call-ffi

# Initialize Go module and download dependencies
echo "📥 Downloading Go dependencies..."
go mod tidy

# Build the Go example
echo "🔨 Building Go example..."
go build -o bsc-call-ffi main.go

echo "✅ Build completed successfully!"
echo "🚀 Run with: ./bsc-call-ffi" 