package main

/*
#cgo LDFLAGS: -L../../../../target/release -lrevm_ffi
#include "../../revm_ffi.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"log"
	"math/big"
	"unsafe"
)

// BSC Testnet configuration
const (
	bscTestnetChainID = 97
)

func main() {
	fmt.Println("🚀 BSC Call FFI Example - With Custom Configuration")

	// Demonstrate different ways to create REVM instances
	demonstrateRevmConfigurations()

	fmt.Println("\n🎉 Configuration demonstration completed!")
}

func demonstrateRevmConfigurations() {
	fmt.Println("\n🔧 Demonstrating REVM Configuration Options...")

	// 1. Default configuration (Ethereum mainnet)
	fmt.Println("\n1️⃣ Creating REVM with default configuration (Ethereum mainnet)...")
	defaultInstance := C.revm_new()
	if defaultInstance == nil {
		log.Fatal("❌ Failed to create default REVM instance")
	}
	defer C.revm_free(defaultInstance)
	
	defaultChainID := C.revm_get_chain_id(defaultInstance)
	defaultSpecID := C.revm_get_spec_id(defaultInstance)
	fmt.Printf("   ✅ Chain ID: %d, Spec ID: %d\n", defaultChainID, defaultSpecID)

	// 2. BSC Testnet preset
	fmt.Println("\n2️⃣ Creating REVM with BSC Testnet preset...")
	bscInstance := C.revm_new_with_preset(C.BSC_TESTNET)
	if bscInstance == nil {
		log.Fatal("❌ Failed to create BSC REVM instance")
	}
	defer C.revm_free(bscInstance)
	
	bscChainID := C.revm_get_chain_id(bscInstance)
	bscSpecID := C.revm_get_spec_id(bscInstance)
	fmt.Printf("   ✅ Chain ID: %d, Spec ID: %d\n", bscChainID, bscSpecID)

	// 3. BSC Mainnet preset
	fmt.Println("\n3️⃣ Creating REVM with BSC Mainnet preset...")
	bscMainnetInstance := C.revm_new_with_preset(C.BSC_MAINNET)
	if bscMainnetInstance == nil {
		log.Fatal("❌ Failed to create BSC Mainnet REVM instance")
	}
	defer C.revm_free(bscMainnetInstance)
	
	bscMainnetChainID := C.revm_get_chain_id(bscMainnetInstance)
	bscMainnetSpecID := C.revm_get_spec_id(bscMainnetInstance)
	fmt.Printf("   ✅ Chain ID: %d, Spec ID: %d\n", bscMainnetChainID, bscMainnetSpecID)

	// 4. Custom configuration
	fmt.Println("\n4️⃣ Creating REVM with custom configuration...")
	
	// Create custom config for BSC testnet with relaxed checks
	config := C.RevmConfigFFI{
		chain_id:                 bscTestnetChainID,
		spec_id:                  18, // Cancun (BSC uses Cancun-equivalent)
		disable_nonce_check:      true,  // Disable for testing
		disable_balance_check:    false, // Keep balance checks
		disable_block_gas_limit:  true,  // Disable for testing
		disable_base_fee:         false, // Keep base fee checks
		max_code_size:            0,     // Use default 24KB limit
	}
	
	customInstance := C.revm_new_with_config(&config)
	if customInstance == nil {
		log.Fatal("❌ Failed to create custom REVM instance")
	}
	defer C.revm_free(customInstance)
	
	customChainID := C.revm_get_chain_id(customInstance)
	customSpecID := C.revm_get_spec_id(customInstance)
	fmt.Printf("   ✅ Chain ID: %d, Spec ID: %d (with relaxed checks)\n", customChainID, customSpecID)

	// 5. Demonstrate account operations with BSC configuration
	fmt.Println("\n5️⃣ Demonstrating account operations with BSC configuration...")
	demonstrateAccountOperations(bscInstance)
}

func demonstrateAccountOperations(instance *C.RevmInstance) {
	// Test addresses
	aliceAddr := "0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6"
	bobAddr := "0x8ba1f109551bD432803012645Hac136c12345678"
	
	fmt.Printf("   👤 Alice: %s\n", aliceAddr)
	fmt.Printf("   👤 Bob: %s\n", bobAddr)

	// Set Alice's balance to 5 ETH
	aliceBalance := "0x4563918244f40000" // 5 ETH in hex
	aliceAddrC := C.CString(aliceAddr)
	aliceBalanceC := C.CString(aliceBalance)
	defer C.free(unsafe.Pointer(aliceAddrC))
	defer C.free(unsafe.Pointer(aliceBalanceC))
	
	if C.revm_set_balance(instance, aliceAddrC, aliceBalanceC) != 0 {
		fmt.Printf("   ❌ Failed to set Alice's balance\n")
		return
	}
	
	// Get Alice's balance
	retrievedBalance := C.revm_get_balance(instance, aliceAddrC)
	if retrievedBalance != nil {
		defer C.revm_free_string(retrievedBalance)
		balanceStr := C.GoString(retrievedBalance)
		fmt.Printf("   💰 Alice's balance set to: %s\n", balanceStr)
		
		// Convert hex to ETH for display
		if balance, ok := new(big.Int).SetString(balanceStr[2:], 16); ok {
			ether := new(big.Float).SetInt(balance)
			ether.Quo(ether, big.NewFloat(1e18))
			fmt.Printf("   💰 Alice's balance: %s ETH\n", ether.Text('f', 6))
		}
	}

	// Set Alice's nonce
	if C.revm_set_nonce(instance, aliceAddrC, 42) != 0 {
		fmt.Printf("   ❌ Failed to set Alice's nonce\n")
		return
	}
	
	// Get Alice's nonce
	nonce := C.revm_get_nonce(instance, aliceAddrC)
	fmt.Printf("   🔢 Alice's nonce: %d\n", nonce)

	// Demonstrate ETH transfer
	fmt.Println("\n   🔄 Performing ETH transfer (Alice → Bob: 1 ETH)...")
	
	bobAddrC := C.CString(bobAddr)
	transferValue := C.CString("0xde0b6b3a7640000") // 1 ETH
	defer C.free(unsafe.Pointer(bobAddrC))
	defer C.free(unsafe.Pointer(transferValue))
	
	transferResult := C.revm_transfer(
		instance,
		aliceAddrC,
		bobAddrC,
		transferValue,
		21000, // Standard gas limit for ETH transfer
	)
	
	if transferResult == nil {
		fmt.Printf("   ❌ Transfer failed\n")
		return
	}
	defer C.revm_free_execution_result(transferResult)
	
	if transferResult.success == 1 {
		fmt.Printf("   ✅ Transfer successful! Gas used: %d\n", transferResult.gas_used)
		
		// Check final balances
		aliceBalanceAfter := C.revm_get_balance(instance, aliceAddrC)
		bobBalanceAfter := C.revm_get_balance(instance, bobAddrC)
		
		if aliceBalanceAfter != nil && bobBalanceAfter != nil {
			defer C.revm_free_string(aliceBalanceAfter)
			defer C.revm_free_string(bobBalanceAfter)
			
			fmt.Printf("   💰 Alice's balance after: %s\n", C.GoString(aliceBalanceAfter))
			fmt.Printf("   💰 Bob's balance after: %s\n", C.GoString(bobBalanceAfter))
		}
	} else {
		fmt.Printf("   ❌ Transfer failed with status: %d\n", transferResult.success)
	}
} 