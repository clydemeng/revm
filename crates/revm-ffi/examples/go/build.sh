#!/bin/bash

# Build script for REVM FFI Go example

set -e

echo "🔨 Building REVM FFI library..."

# Navigate to the root directory
cd ../../../../

# Build the FFI library in release mode
cargo build --release -p revm-ffi

if [ $? -ne 0 ]; then
    echo "❌ Failed to build REVM FFI library"
    exit 1
fi

echo "✅ REVM FFI library built successfully"

# Copy the header file to the Go example directory
cp crates/revm-ffi/revm_ffi.h crates/revm-ffi/examples/go/

echo "📋 Header file copied to Go example directory"

# Navigate to the Go example directory
cd crates/revm-ffi/examples/go

# Set the library path for the current platform
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "🍎 macOS detected - set DYLD_LIBRARY_PATH"
    export DYLD_LIBRARY_PATH="$(pwd)/target/release:$DYLD_LIBRARY_PATH"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "🐧 Linux detected - set LD_LIBRARY_PATH"
    export LD_LIBRARY_PATH="$(pwd)/target/release:$LD_LIBRARY_PATH"
else
    echo "❓ Unknown OS, assuming Linux-like"
    export LD_LIBRARY_PATH="$(pwd)/target/release:$LD_LIBRARY_PATH"
fi

echo "🚀 Running Go example..."

# Run the Go example
go run main.go

echo "🎉 Go example completed successfully!" 