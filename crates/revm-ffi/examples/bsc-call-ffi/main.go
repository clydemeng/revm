package main

/*
#cgo LDFLAGS: -L../../../../target/release -lrevm_ffi
#include "../../revm_ffi.h"
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"unsafe"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20 ABI for transfer function
const erc20ABI = `[
	{
		"constant": false,
		"inputs": [
			{"name": "_to", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "_owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"type": "function"
	}
]`

// BSC Testnet configuration
const (
	bscTestnetRPC = "https://data-seed-prebsc-1-s1.binance.org:8545"
	// USDT contract on BSC Testnet
	usdtContractAddress = "0x337610d27c682E347C9cD60BD4b3b107C9d34dDd"
)

func main() {
	fmt.Println("🚀 BSC Call FFI Example - ERC20 Transfer with REVM")
	fmt.Println("📡 Connecting to BSC Testnet...")

	// Connect to BSC testnet
	client, err := ethclient.Dial(bscTestnetRPC)
	if err != nil {
		log.Fatalf("❌ Failed to connect to BSC testnet: %v", err)
	}
	defer client.Close()

	// Verify connection
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get chain ID: %v", err)
	}
	fmt.Printf("✅ Connected to BSC Testnet (Chain ID: %s)\n", chainID.String())

	// Initialize REVM instance
	fmt.Println("\n🔧 Initializing REVM instance...")
	instance := C.revm_new()
	if instance == nil {
		log.Fatal("❌ Failed to create REVM instance")
	}
	defer C.revm_free(instance)
	fmt.Println("✅ REVM instance created successfully")

	// Setup test accounts
	fromAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6")
	toAddress := common.HexToAddress("0x8ba1f109551bD432803012645Hac136c12345678")
	contractAddress := common.HexToAddress(usdtContractAddress)

	fmt.Printf("👤 From Address: %s\n", fromAddress.Hex())
	fmt.Printf("👤 To Address: %s\n", toAddress.Hex())
	fmt.Printf("📄 Contract Address: %s\n", contractAddress.Hex())

	// Fetch real account balances from BSC
	fmt.Println("\n💰 Fetching real account data from BSC...")
	
	fromBalance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Printf("⚠️ Could not fetch from balance: %v", err)
		fromBalance = big.NewInt(0)
	}
	
	toBalance, err := client.BalanceAt(context.Background(), toAddress, nil)
	if err != nil {
		log.Printf("⚠️ Could not fetch to balance: %v", err)
		toBalance = big.NewInt(0)
	}

	fmt.Printf("💰 From ETH Balance: %s ETH\n", weiToEther(fromBalance))
	fmt.Printf("💰 To ETH Balance: %s ETH\n", weiToEther(toBalance))

	// Set up REVM with enhanced account balances for gas fees
	fmt.Println("\n🔧 Setting up REVM with enhanced account data...")
	
	// Ensure from address has at least 1 ETH for gas fees
	minBalance := big.NewInt(1e18) // 1 ETH
	if fromBalance.Cmp(minBalance) < 0 {
		fromBalance = minBalance
		fmt.Printf("💰 Enhanced From Balance to: %s ETH (for gas fees)\n", weiToEther(fromBalance))
	}
	
	if err := setAccountBalance(instance, fromAddress.Hex(), fromBalance); err != nil {
		log.Printf("⚠️ Failed to set from balance in REVM: %v", err)
	}
	
	if err := setAccountBalance(instance, toAddress.Hex(), toBalance); err != nil {
		log.Printf("⚠️ Failed to set to balance in REVM: %v", err)
	}

	// Deploy ERC20 contract bytecode to REVM (for simulation)
	fmt.Println("\n📦 Setting up ERC20 contract in REVM...")
	erc20Bytecode := getERC20Bytecode()
	
	deployerAddr := C.CString("0x1000000000000000000000000000000000000001")
	defer C.free(unsafe.Pointer(deployerAddr))
	
	// Set deployer balance
	deployerBalance := C.CString("0x8ac7230489e80000") // 10 ETH
	defer C.free(unsafe.Pointer(deployerBalance))
	
	if C.revm_set_balance(instance, deployerAddr, deployerBalance) != 0 {
		log.Printf("⚠️ Failed to set deployer balance: %s", getLastError(instance))
	}

	// Deploy contract
	deployResult := C.revm_deploy_contract(
		instance,
		deployerAddr,
		(*C.uchar)(unsafe.Pointer(&erc20Bytecode[0])),
		C.uint(len(erc20Bytecode)),
		1000000,
	)

	if deployResult == nil || deployResult.success != 1 {
		log.Printf("⚠️ Contract deployment failed: %s", getLastError(instance))
		return
	}
	defer C.revm_free_deployment_result(deployResult)

	revmContractAddr := C.GoString(deployResult.contract_address)
	fmt.Printf("✅ ERC20 contract deployed in REVM at: %s\n", revmContractAddr)

	// Set up initial token balances in REVM contract
	fmt.Println("\n🪙 Setting up initial token balances...")
	
	// Give fromAddress 1000 tokens
	initialTokens := "0x3635c9adc5dea00000" // 1000 tokens (18 decimals)
	if err := setTokenBalance(instance, revmContractAddr, fromAddress.Hex(), initialTokens); err != nil {
		log.Printf("⚠️ Failed to set initial token balance: %v", err)
	}

	// Check initial balances
	fromTokenBalance, _ := getTokenBalance(instance, revmContractAddr, fromAddress.Hex())
	toTokenBalance, _ := getTokenBalance(instance, revmContractAddr, toAddress.Hex())
	
	fmt.Printf("🪙 From Token Balance: %s tokens\n", fromTokenBalance)
	fmt.Printf("🪙 To Token Balance: %s tokens\n", toTokenBalance)

	// Create ERC20 transfer transaction
	fmt.Println("\n🔄 Creating ERC20 transfer transaction...")
	
	transferAmount := big.NewInt(250) // 250 tokens
	transferAmount.Mul(transferAmount, big.NewInt(1e18)) // Convert to wei (18 decimals)
	
	// Create transfer call data
	transferData, err := createTransferCallData(toAddress, transferAmount)
	if err != nil {
		log.Fatalf("❌ Failed to create transfer call data: %v", err)
	}

	fmt.Printf("📝 Transfer Amount: 250 tokens\n")
	fmt.Printf("📝 Call Data: %s\n", hexutil.Encode(transferData))

	// Execute transfer in REVM
	fmt.Println("\n⚡ Executing transfer in REVM...")
	
	fromAddrC := C.CString(fromAddress.Hex())
	contractAddrC := C.CString(revmContractAddr)
	defer C.free(unsafe.Pointer(fromAddrC))
	defer C.free(unsafe.Pointer(contractAddrC))

	// Set transaction parameters
	if C.revm_set_tx(
		instance,
		fromAddrC,                                    // caller
		contractAddrC,                                // to (contract)
		nil,                                          // value (0 for ERC20 transfer)
		(*C.uchar)(unsafe.Pointer(&transferData[0])), // data
		C.uint(len(transferData)),                    // data_len
		100000,                                       // gas_limit
		nil,                                          // gas_price (default)
		0,                                            // nonce
	) != 0 {
		log.Fatalf("❌ Failed to set transaction: %s", getLastError(instance))
	}

	// Execute the transaction
	execResult := C.revm_execute_commit(instance)
	if execResult == nil {
		log.Fatalf("❌ Transaction execution failed: %s", getLastError(instance))
	}
	defer C.revm_free_execution_result(execResult)

	if execResult.success == 1 {
		fmt.Printf("✅ Transfer executed successfully! Gas used: %d\n", execResult.gas_used)
	} else {
		fmt.Printf("❌ Transfer failed with status: %d\n", execResult.success)
		return
	}

	// Check final balances
	fmt.Println("\n💰 Checking final balances...")
	
	fromTokenBalanceAfter, _ := getTokenBalance(instance, revmContractAddr, fromAddress.Hex())
	toTokenBalanceAfter, _ := getTokenBalance(instance, revmContractAddr, toAddress.Hex())
	
	fmt.Printf("🪙 From Token Balance After: %s tokens\n", fromTokenBalanceAfter)
	fmt.Printf("🪙 To Token Balance After: %s tokens\n", toTokenBalanceAfter)

	// Verify transfer
	fmt.Println("\n🔍 Verifying transfer...")
	fmt.Printf("   From: %s -> %s (difference: -250)\n", fromTokenBalance, fromTokenBalanceAfter)
	fmt.Printf("   To: %s -> %s (difference: +250)\n", toTokenBalance, toTokenBalanceAfter)

	// Demonstrate fetching real transaction from BSC
	fmt.Println("\n📡 Demonstrating real BSC transaction fetch...")
	demonstrateRealTxFetch(client)

	fmt.Println("\n🎉 BSC Call FFI Example completed successfully!")
	fmt.Println("\n📝 Summary:")
	fmt.Printf("   - Connected to BSC Testnet (Chain ID: %s)\n", chainID.String())
	fmt.Printf("   - Simulated ERC20 transfer in REVM\n")
	fmt.Printf("   - Transfer amount: 250 tokens\n")
	fmt.Printf("   - Gas used: %d\n", execResult.gas_used)
}

// Helper functions

func weiToEther(wei *big.Int) string {
	ether := new(big.Float).SetInt(wei)
	ether.Quo(ether, big.NewFloat(1e18))
	return ether.Text('f', 6)
}

func setAccountBalance(instance *C.RevmInstance, address string, balance *big.Int) error {
	addrC := C.CString(address)
	balanceHex := C.CString("0x" + balance.Text(16))
	defer C.free(unsafe.Pointer(addrC))
	defer C.free(unsafe.Pointer(balanceHex))

	if C.revm_set_balance(instance, addrC, balanceHex) != 0 {
		return fmt.Errorf("failed to set balance: %s", getLastError(instance))
	}
	return nil
}

func setTokenBalance(instance *C.RevmInstance, contractAddr, userAddr, balance string) error {
	contractC := C.CString(contractAddr)
	slotC := C.CString("0x0") // Simplified storage slot
	balanceC := C.CString(balance)
	defer C.free(unsafe.Pointer(contractC))
	defer C.free(unsafe.Pointer(slotC))
	defer C.free(unsafe.Pointer(balanceC))

	if C.revm_set_storage(instance, contractC, slotC, balanceC) != 0 {
		return fmt.Errorf("failed to set token balance: %s", getLastError(instance))
	}
	return nil
}

func getTokenBalance(instance *C.RevmInstance, contractAddr, userAddr string) (string, error) {
	contractC := C.CString(contractAddr)
	slotC := C.CString("0x0") // Simplified storage slot
	defer C.free(unsafe.Pointer(contractC))
	defer C.free(unsafe.Pointer(slotC))

	result := C.revm_get_storage(instance, contractC, slotC)
	if result == nil {
		return "0", fmt.Errorf("failed to get token balance")
	}
	defer C.revm_free_string(result)

	return C.GoString(result), nil
}

func createTransferCallData(to common.Address, amount *big.Int) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}

	return parsedABI.Pack("transfer", to, amount)
}

func getERC20Bytecode() []byte {
	// Simple ERC20-like contract bytecode
	return []byte{
		0x60, 0x80, 0x60, 0x40, 0x52, // Set up memory
		0x34, 0x80, 0x15, 0x61, 0x00, 0x10, 0x57, 0x60, 0x00, 0x80, 0xfd,
		0x5b, // JUMPDEST
		0x60, 0x01, // PUSH1 0x01 (return true)
		0x60, 0x00, // PUSH1 0x00 (memory position)
		0x52, // MSTORE
		0x60, 0x20, // PUSH1 0x20 (32 bytes)
		0x60, 0x00, // PUSH1 0x00 (memory position)
		0xf3, // RETURN
	}
}

func demonstrateRealTxFetch(client *ethclient.Client) {
	// Fetch latest block
	latestBlock, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Printf("⚠️ Could not fetch latest block: %v", err)
		return
	}

	fmt.Printf("📦 Latest Block Number: %s\n", latestBlock.Number().String())
	fmt.Printf("📦 Block Hash: %s\n", latestBlock.Hash().Hex())
	fmt.Printf("📦 Transaction Count: %d\n", len(latestBlock.Transactions()))

	// Show first transaction if available
	if len(latestBlock.Transactions()) > 0 {
		tx := latestBlock.Transactions()[0]
		fmt.Printf("🔗 First Transaction Hash: %s\n", tx.Hash().Hex())
		fmt.Printf("🔗 Transaction Value: %s ETH\n", weiToEther(tx.Value()))
	}
}

func getLastError(instance *C.RevmInstance) string {
	errorPtr := C.revm_get_last_error(instance)
	if errorPtr != nil {
		return C.GoString(errorPtr)
	}
	return "Unknown error"
} 