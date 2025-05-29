# REVM FFI

This crate provides Foreign Function Interface (FFI) bindings for REVM, allowing other languages like Go to interact with the Rust Ethereum Virtual Machine through C-compatible APIs.

## Features

- **Complete EVM functionality**: Execute transactions, deploy contracts, manage state
- **Memory-safe**: Proper memory management with explicit cleanup functions
- **Error handling**: Comprehensive error reporting through C-compatible interfaces
- **Go integration**: Ready-to-use with CGO for seamless Go integration
- **Cross-platform**: Works on Linux, macOS, and Windows

## Building

### Prerequisites

- Rust toolchain (1.70+)
- C compiler (GCC, Clang, or MSVC)
- Go 1.19+ (for Go examples)

### Build the FFI Library

```bash
# Build the release version
cargo build --release -p revm-ffi

# The library will be available at:
# - Linux/macOS: target/release/librevm_ffi.so / target/release/librevm_ffi.dylib
# - Windows: target/release/revm_ffi.dll
```

### Build Static Library

```bash
# For static linking
cargo build --release -p revm-ffi
# Static library: target/release/librevm_ffi.a
```

## Usage with Go

### 1. Copy Header File

Copy the header file to your Go project:

```bash
cp crates/revm-ffi/revm_ffi.h /path/to/your/go/project/
```

### 2. Go Integration

```go
package main

/*
#cgo LDFLAGS: -L/path/to/revm/target/release -lrevm_ffi
#include "revm_ffi.h"
#include <stdlib.h>
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    // Initialize REVM instance
    instance := C.revm_new()
    if instance == nil {
        panic("Failed to create REVM instance")
    }
    defer C.revm_free(instance)

    // Set account balance
    address := C.CString("0x1000000000000000000000000000000000000001")
    balance := C.CString("0x3635c9adc5dea00000") // 1000 ETH
    defer C.free(unsafe.Pointer(address))
    defer C.free(unsafe.Pointer(balance))

    if C.revm_set_balance(instance, address, balance) != 0 {
        fmt.Printf("Error: %s\n", getLastError(instance))
        return
    }

    // Deploy contract
    bytecode := []byte{0x60, 0x80, 0x60, 0x40, 0x52} // Simple bytecode
    result := C.revm_deploy_contract(
        instance,
        address,
        (*C.uchar)(unsafe.Pointer(&bytecode[0])),
        C.uint(len(bytecode)),
        1000000,
    )
    
    if result != nil {
        defer C.revm_free_deployment_result(result)
        if result.success == 1 {
            fmt.Printf("Contract deployed at: %s\n", C.GoString(result.contract_address))
        }
    }
}

func getLastError(instance *C.RevmInstance) string {
    errorPtr := C.revm_get_last_error(instance)
    if errorPtr != nil {
        return C.GoString(errorPtr)
    }
    return "Unknown error"
}
```

### 3. Running the Example

```bash
# Navigate to the Go example directory
cd crates/revm-ffi/examples/go

# Build the REVM FFI library first
cd ../../../..
cargo build --release -p revm-ffi

# Run the Go example
cd crates/revm-ffi/examples/go
go run main.go
```

## API Reference

### Core Functions

#### `revm_new()`
Creates a new REVM instance.
- **Returns**: Pointer to REVM instance or NULL on failure

#### `revm_free(instance)`
Frees a REVM instance and all associated resources.
- **Parameters**: `instance` - REVM instance pointer

#### `revm_set_tx(...)`
Sets transaction parameters for execution.
- **Parameters**:
  - `instance` - REVM instance
  - `caller` - Caller address (hex string)
  - `to` - Recipient address (hex string, NULL for contract creation)
  - `value` - Transaction value (hex string, NULL for 0)
  - `data` - Transaction data bytes
  - `data_len` - Length of transaction data
  - `gas_limit` - Gas limit
  - `gas_price` - Gas price (hex string, NULL for default)
  - `nonce` - Transaction nonce
- **Returns**: 0 on success, -1 on failure

#### `revm_execute(instance)`
Executes a transaction without committing state changes.
- **Returns**: Execution result or NULL on failure

#### `revm_execute_commit(instance)`
Executes and commits a transaction.
- **Returns**: Execution result or NULL on failure

### Account Management

#### `revm_get_balance(instance, address)`
Gets account balance.
- **Returns**: Balance as hex string or NULL on failure

#### `revm_set_balance(instance, address, balance)`
Sets account balance.
- **Returns**: 0 on success, -1 on failure

### Storage Operations

#### `revm_get_storage(instance, address, slot)`
Gets storage value from a contract.
- **Returns**: Storage value as hex string or NULL on failure

#### `revm_set_storage(instance, address, slot, value)`
Sets storage value in a contract.
- **Returns**: 0 on success, -1 on failure

### Contract Deployment

#### `revm_deploy_contract(instance, deployer, bytecode, bytecode_len, gas_limit)`
Deploys a smart contract.
- **Returns**: Deployment result or NULL on failure

### Error Handling

#### `revm_get_last_error(instance)`
Gets the last error message.
- **Returns**: Error string or NULL if no error

### Memory Management

#### `revm_free_string(s)`
Frees a string allocated by the library.

#### `revm_free_execution_result(result)`
Frees an execution result structure.

#### `revm_free_deployment_result(result)`
Frees a deployment result structure.

## Data Structures

### ExecutionResultFFI
```c
typedef struct {
    int success;           // 1 = success, 0 = revert, -1 = halt
    unsigned int gas_used;
    unsigned int gas_refunded;
    unsigned char* output_data;
    unsigned int output_len;
    unsigned int logs_count;
    struct LogFFI* logs;
    char* created_address;  // Only for contract creation
} ExecutionResultFFI;
```

### DeploymentResultFFI
```c
typedef struct {
    int success;
    char* contract_address;
    unsigned int gas_used;
    unsigned int gas_refunded;
} DeploymentResultFFI;
```

### LogFFI
```c
typedef struct LogFFI {
    char* address;
    unsigned int topics_count;
    char** topics;
    unsigned char* data;
    unsigned int data_len;
} LogFFI;
```

## Best Practices

1. **Always check return values** for NULL pointers and error codes
2. **Free allocated memory** using the provided cleanup functions
3. **Handle errors gracefully** by checking `revm_get_last_error()`
4. **Use proper hex formatting** for addresses and values (with 0x prefix)
5. **Set appropriate gas limits** to avoid out-of-gas errors

## Thread Safety

The FFI is **not thread-safe**. Each thread should use its own REVM instance. Do not share instances across threads without proper synchronization.

## Performance Considerations

- Reuse REVM instances when possible
- Batch operations to minimize FFI overhead
- Use appropriate gas limits to avoid unnecessary computation
- Free resources promptly to avoid memory leaks

## Troubleshooting

### Common Issues

1. **Library not found**: Ensure the library path is correct in CGO LDFLAGS
2. **Invalid addresses**: Use proper 40-character hex addresses with 0x prefix
3. **Memory leaks**: Always call the corresponding free functions
4. **Gas errors**: Set appropriate gas limits for operations

### Debug Build

For debugging, build with debug symbols:

```bash
cargo build -p revm-ffi
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 