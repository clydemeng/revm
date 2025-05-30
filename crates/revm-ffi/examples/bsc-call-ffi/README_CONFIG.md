# REVM FFI Configuration Guide

This guide demonstrates the enhanced REVM FFI interface that supports custom chain configurations, including BSC (Binance Smart Chain) specific settings.

## Overview

The enhanced REVM FFI interface now supports:

- **Chain-specific configurations** (Ethereum, BSC Mainnet, BSC Testnet)
- **Custom hardfork specifications** (Frontier through Prague)
- **Configurable validation rules** (nonce checks, balance checks, gas limits)
- **BSC-optimized settings** for better compatibility

## Configuration Options

### 1. Default Configuration (Ethereum Mainnet)

```go
// Creates REVM with Ethereum mainnet defaults
instance := C.revm_new()
// Chain ID: 1, Spec: Prague (latest)
```

### 2. Predefined Chain Presets

```go
// BSC Testnet (Chapel)
bscTestnet := C.revm_new_with_preset(C.BSC_TESTNET)
// Chain ID: 97, Spec: Cancun

// BSC Mainnet
bscMainnet := C.revm_new_with_preset(C.BSC_MAINNET)
// Chain ID: 56, Spec: Cancun

// Ethereum Mainnet
ethMainnet := C.revm_new_with_preset(C.ETHEREUM_MAINNET)
// Chain ID: 1, Spec: Prague
```

### 3. Custom Configuration

```go
config := C.RevmConfigFFI{
    chain_id:                 97,    // BSC Testnet
    spec_id:                  18,    // Cancun hardfork
    disable_nonce_check:      true,  // Disable for testing
    disable_balance_check:    false, // Keep balance validation
    disable_block_gas_limit:  true,  // Disable for simulation
    disable_base_fee:         false, // Keep EIP-1559 validation
    max_code_size:            0,     // Use default 24KB limit
}

customInstance := C.revm_new_with_config(&config)
```

## Hardfork Specifications

The `spec_id` field supports all Ethereum hardforks:

| Spec ID | Hardfork Name | Description |
|---------|---------------|-------------|
| 0 | Frontier | Original Ethereum |
| 1 | Frontier Thawing | Ice age delay |
| 2 | Homestead | First major upgrade |
| 3 | DAO Fork | DAO incident response |
| 4 | Tangerine Whistle | Gas cost adjustments |
| 5 | Spurious Dragon | State trie clearing |
| 6 | Byzantium | Privacy and scaling |
| 7 | Constantinople | Efficiency improvements |
| 8 | Petersburg | Constantinople fix |
| 9 | Istanbul | Gas optimizations |
| 10 | Muir Glacier | Ice age delay |
| 11 | Berlin | Gas cost changes |
| 12 | London | EIP-1559 fee market |
| 13 | Arrow Glacier | Ice age delay |
| 14 | Gray Glacier | Ice age delay |
| 15 | Merge | Proof of Stake transition |
| 16 | Shanghai | Withdrawals enabled |
| 17 | Cancun | Blob transactions |
| 18 | Prague | Latest features |

## BSC-Specific Considerations

### Chain IDs
- **BSC Mainnet**: 56
- **BSC Testnet (Chapel)**: 97

### Recommended Settings for BSC
```go
bscConfig := C.RevmConfigFFI{
    chain_id:                 56,    // or 97 for testnet
    spec_id:                  17,    // Cancun (BSC is typically one behind)
    disable_nonce_check:      false, // Keep nonce validation
    disable_balance_check:    false, // Keep balance validation
    disable_block_gas_limit:  false, // Keep gas limit validation
    disable_base_fee:         false, // BSC supports EIP-1559
    max_code_size:            0,     // Use default limit
}
```

### BSC vs Ethereum Differences
- **Gas Price**: BSC uses lower gas prices (typically 5 Gwei vs 20+ Gwei)
- **Block Time**: BSC has ~3 second blocks vs Ethereum's ~12 seconds
- **Hardforks**: BSC typically implements hardforks slightly after Ethereum
- **Consensus**: BSC uses Proof of Staked Authority (PoSA) vs Ethereum's Proof of Stake

## Configuration Queries

You can query the configuration of any REVM instance:

```go
chainID := C.revm_get_chain_id(instance)
specID := C.revm_get_spec_id(instance)

fmt.Printf("Chain ID: %d, Spec ID: %d\n", chainID, specID)
```

## Account Management

The enhanced interface provides comprehensive account management:

```go
// Set account balance
C.revm_set_balance(instance, addressC, balanceC)

// Get account balance
balance := C.revm_get_balance(instance, addressC)

// Set account nonce
C.revm_set_nonce(instance, addressC, 42)

// Get account nonce
nonce := C.revm_get_nonce(instance, addressC)
```

## Transaction Operations

### ETH Transfers
```go
result := C.revm_transfer(
    instance,
    fromAddressC,
    toAddressC,
    valueC,        // Amount in wei (hex string)
    21000,         // Gas limit
)
```

### Contract Calls
```go
result := C.revm_call_contract(
    instance,
    fromAddressC,
    contractAddressC,
    callDataPtr,   // ABI-encoded function call
    callDataLen,
    valueC,        // ETH value to send (can be "0x0")
    gasLimit,
)
```

### Contract Deployment
```go
deployResult := C.revm_deploy_contract(
    instance,
    deployerAddressC,
    bytecodePtr,
    bytecodeLen,
    gasLimit,
)
```

## Example Usage

See `config_demo.go` for a complete example that demonstrates:

1. Creating REVM instances with different configurations
2. Querying chain and spec information
3. Setting up accounts with balances and nonces
4. Performing ETH transfers
5. Error handling and memory management

## Building and Running

1. **Build the FFI library**:
   ```bash
   cargo build --release -p revm-ffi
   ```

2. **Run the configuration demo**:
   ```bash
   go run config_demo.go
   ```

3. **Run the full BSC example**:
   ```bash
   go run main.go
   ```

## Memory Management

Always remember to free allocated resources:

```go
defer C.revm_free(instance)
defer C.revm_free_string(stringResult)
defer C.revm_free_execution_result(execResult)
defer C.revm_free_deployment_result(deployResult)
defer C.free(unsafe.Pointer(cString))
```

## Error Handling

Check for errors and retrieve error messages:

```go
if result == nil {
    errorMsg := C.revm_get_last_error(instance)
    if errorMsg != nil {
        fmt.Printf("Error: %s\n", C.GoString(errorMsg))
    }
}
```

## Integration with Real Networks

The REVM FFI can be used alongside real blockchain clients:

1. **Fetch real state** from BSC/Ethereum nodes
2. **Simulate transactions** in REVM with that state
3. **Compare results** between simulation and actual execution
4. **Test contract behavior** before deployment

This makes it ideal for:
- **Transaction simulation**
- **Gas estimation**
- **Contract testing**
- **MEV analysis**
- **DeFi strategy backtesting**

## Conclusion

The enhanced REVM FFI interface provides a powerful and flexible way to simulate Ethereum and BSC transactions with full control over chain parameters, validation rules, and execution environment. This makes it suitable for a wide range of blockchain development and analysis tasks. 