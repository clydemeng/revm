# BSC Call FFI Example

This example demonstrates how to use REVM FFI bindings with Binance Smart Chain (BSC) to simulate ERC20 token transfers. It combines real blockchain data fetching with REVM's high-performance EVM simulation.

## Features

- 🌐 **BSC Integration**: Connects to BSC Testnet to fetch real account balances and block data
- 🔗 **FFI Bindings**: Uses REVM's C FFI interface from Go
- 🪙 **ERC20 Simulation**: Simulates ERC20 token transfers using REVM
- ⚡ **High Performance**: Leverages REVM's optimized EVM implementation
- 📊 **Real Data**: Fetches actual account balances and blockchain state

## What This Example Does

1. **Connects to BSC Testnet** - Establishes connection to Binance Smart Chain testnet
2. **Fetches Real Data** - Gets actual account balances and latest block information
3. **Initializes REVM** - Creates a REVM instance via FFI
4. **Sets Up Accounts** - Configures test accounts with real balance data
5. **Deploys ERC20 Contract** - Deploys a simple ERC20-like contract in REVM
6. **Executes Transfer** - Simulates an ERC20 token transfer (250 tokens)
7. **Verifies Results** - Checks balances before and after transfer
8. **Demonstrates Integration** - Shows how to combine real blockchain data with REVM simulation

## Prerequisites

- **Rust** (for building REVM FFI library)
- **Go 1.21+** (for running the example)
- **Internet connection** (for BSC testnet access)

## Building and Running

### Option 1: Using the build script (recommended)

```bash
./build.sh
./bsc-call-ffi
```

### Option 2: Manual build

```bash
# Build REVM FFI library
cd ../../../../
cargo build --release -p revm-ffi

# Return to example directory
cd crates/revm-ffi/examples/bsc-call-ffi

# Download Go dependencies
go mod tidy

# Build and run
go build -o bsc-call-ffi main.go
./bsc-call-ffi
```

## Expected Output

```
🚀 BSC Call FFI Example - ERC20 Transfer with REVM
📡 Connecting to BSC Testnet...
✅ Connected to BSC Testnet (Chain ID: 97)

🔧 Initializing REVM instance...
✅ REVM instance created successfully
👤 From Address: 0x742d35cc6634c0532925a3b8d4c9db96c4b4d8b6
👤 To Address: 0x8ba1f109551bd432803012645hac136c
📄 Contract Address: 0x337610d27c682e347c9cd60bd4b3b107c9d34ddd

💰 Fetching real account data from BSC...
💰 From ETH Balance: 0.000000 ETH
💰 To ETH Balance: 0.000000 ETH

🔧 Setting up REVM with real account data...

📦 Setting up ERC20 contract in REVM...
✅ ERC20 contract deployed in REVM at: 0x...

🪙 Setting up initial token balances...
🪙 From Token Balance: 0x3635c9adc5dea00000 tokens
🪙 To Token Balance: 0x0 tokens

🔄 Creating ERC20 transfer transaction...
📝 Transfer Amount: 250 tokens
📝 Call Data: 0xa9059cbb...

⚡ Executing transfer in REVM...
✅ Transfer executed successfully! Gas used: 21000

💰 Checking final balances...
🪙 From Token Balance After: 0x... tokens
🪙 To Token Balance After: 0x... tokens

🔍 Verifying transfer...
   From: 0x3635c9adc5dea00000 -> 0x... (difference: -250)
   To: 0x0 -> 0x... (difference: +250)

📡 Demonstrating real BSC transaction fetch...
📦 Latest Block Number: 12345678
📦 Block Hash: 0x...
📦 Transaction Count: 42
🔗 First Transaction Hash: 0x...
🔗 Transaction Value: 0.100000 ETH

🎉 BSC Call FFI Example completed successfully!

📝 Summary:
   - Connected to BSC Testnet (Chain ID: 97)
   - Simulated ERC20 transfer in REVM
   - Transfer amount: 250 tokens
   - Gas used: 21000
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Client     │    │   REVM FFI      │    │   REVM Core     │
│                 │    │                 │    │                 │
│ • BSC Client    │───▶│ • C Interface   │───▶│ • EVM Engine    │
│ • ABI Encoding  │    │ • Memory Mgmt   │    │ • State Mgmt    │
│ • Data Fetching │    │ • Type Conv     │    │ • Execution     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              │
         ▼                                              ▼
┌─────────────────┐                            ┌─────────────────┐
│   BSC Testnet   │                            │   Simulated     │
│                 │                            │   EVM State     │
│ • Real Balances │                            │                 │
│ • Block Data    │                            │ • Accounts      │
│ • Transactions  │                            │ • Contracts     │
└─────────────────┘                            │ • Storage       │
                                               └─────────────────┘
```

## Key Components

### BSC Integration
- **RPC Client**: Connects to BSC testnet via JSON-RPC
- **Real Data**: Fetches actual account balances and block information
- **Chain ID**: Verifies connection to correct network (BSC Testnet = 97)

### REVM FFI Usage
- **Instance Management**: Creates and manages REVM instances
- **Account Setup**: Sets account balances from real BSC data
- **Contract Deployment**: Deploys ERC20 contract bytecode
- **Transaction Execution**: Executes ERC20 transfers with gas tracking

### ERC20 Simulation
- **ABI Encoding**: Properly encodes transfer function calls
- **Token Balances**: Manages token balances in contract storage
- **Transfer Logic**: Simulates token transfers between accounts
- **Balance Verification**: Checks balances before and after transfers

## Configuration

The example uses BSC Testnet by default:
- **RPC URL**: `https://data-seed-prebsc-1-s1.binance.org:8545`
- **Chain ID**: 97 (BSC Testnet)
- **Test Contract**: USDT on BSC Testnet

To use BSC Mainnet, change:
```go
const (
    bscMainnetRPC = "https://bsc-dataseed.binance.org/"
    // Chain ID will be 56 for mainnet
)
```

## Use Cases

This example demonstrates several important use cases:

1. **Transaction Simulation**: Test transactions before sending to real network
2. **Gas Estimation**: Accurate gas usage calculation for ERC20 transfers
3. **State Forking**: Use real blockchain state as starting point for simulations
4. **Integration Testing**: Test dApp logic with real blockchain data
5. **Performance Analysis**: Benchmark EVM execution performance

## Comparison with Other Examples

| Feature | BSC Call FFI | Go FFI | Rust ERC20 |
|---------|-------------|---------|------------|
| Language | Go | Go | Rust |
| Blockchain Data | ✅ Real BSC | ❌ Mock | ❌ Mock |
| ERC20 Transfers | ✅ Full ABI | ❌ ETH only | ✅ Full ERC20 |
| FFI Usage | ✅ Yes | ✅ Yes | ❌ Native |
| Network Connection | ✅ BSC Testnet | ❌ None | ❌ None |
| Real-world Ready | ✅ Yes | ❌ Demo only | ❌ Demo only |

## Troubleshooting

### Build Issues
- Ensure REVM FFI library is built: `cargo build --release -p revm-ffi`
- Check Go version: `go version` (requires 1.21+)
- Verify CGO is enabled: `echo $CGO_ENABLED` (should be 1)

### Network Issues
- Check internet connection for BSC testnet access
- Try alternative BSC RPC endpoints if connection fails
- Verify firewall settings allow HTTPS connections

### Runtime Issues
- Check that `librevm_ffi.so` (or `.dylib` on macOS) exists in `../../../../target/release/`
- Ensure proper library path in CGO LDFLAGS
- Verify account addresses are valid Ethereum addresses

## Next Steps

This example can be extended to:
- Support multiple ERC20 tokens
- Implement more complex DeFi operations
- Add transaction batching
- Support custom contract deployments
- Integrate with other blockchain networks 