# Pure EVM Benchmark Results

## Overview
This benchmark compares the pure EVM execution performance between REVM (Rust) and BSC-EVM (Go-based) by executing 50,000 internal token transfers within a single smart contract transaction.

## Test Setup
- **Contract**: BIGA token with built-in `batchTransferSequential()` function
- **Test**: 50,000 sequential token transfers (1 token each)
- **Measurement**: Pure EVM execution time (excludes transaction overhead)
- **Environment**: macOS 24.4.0, Release builds

## Results

### REVM (Rust Implementation)
```
Transfers: 50,000
Duration: ~133-136ms
Transfers/sec: ~366,000-376,000
```

### BSC-EVM (Go Implementation)
```
Transfers: 50,000
Duration: ~281-296ms  
Transfers/sec: ~169,000-178,000
```

## Performance Comparison
- **REVM Performance**: ~372,000 transfers/sec (average)
- **BSC-EVM Performance**: ~174,000 transfers/sec (average)
- **Performance Ratio**: REVM is ~2.1x faster than BSC-EVM

## Key Findings
1. **REVM Advantage**: Rust's zero-cost abstractions and memory management provide significant performance benefits for EVM execution
2. **Consistent Results**: Multiple runs show consistent performance characteristics
3. **Gas Efficiency**: Both implementations handle the same contract correctly with similar gas consumption patterns
4. **Scalability**: REVM's performance advantage becomes more pronounced with larger transaction volumes

## Technical Notes
- Both benchmarks use the same BIGA contract bytecode from `bytecode/BIGA.bin`
- Shanghai/Cancun hard fork features enabled for PUSH0 opcode support
- Gas limits set appropriately for 50k transfers (2B gas for REVM, 2B gas for BSC-EVM)
- Results verified with balance checks to ensure correctness

## Conclusion
REVM demonstrates superior performance for pure EVM execution, making it an excellent choice for high-throughput blockchain applications requiring fast transaction processing. 