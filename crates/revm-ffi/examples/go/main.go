package main

/*
#cgo LDFLAGS: -L../../../../target/release -lrevm_ffi
#include "../../revm_ffi.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("🚀 REVM FFI Go Example")

	// Initialize REVM instance
	instance := C.revm_new()
	if instance == nil {
		fmt.Println("❌ Failed to create REVM instance")
		return
	}
	defer C.revm_free(instance)

	fmt.Println("✅ REVM instance created successfully")

	// Set up accounts with balances
	deployer := C.CString("0x1000000000000000000000000000000000000001")
	alice := C.CString("0x2000000000000000000000000000000000000002")
	bob := C.CString("0x3000000000000000000000000000000000000003")
	defer C.free(unsafe.Pointer(deployer))
	defer C.free(unsafe.Pointer(alice))
	defer C.free(unsafe.Pointer(bob))

	// Set Alice's balance to 1000 ETH
	aliceBalance := C.CString("0x3635c9adc5dea00000") // 1000 ETH in wei
	defer C.free(unsafe.Pointer(aliceBalance))

	if C.revm_set_balance(instance, alice, aliceBalance) != 0 {
		fmt.Printf("❌ Failed to set Alice's balance: %s\n", getLastError(instance))
		return
	}
	fmt.Println("💰 Set Alice's balance to 1000 ETH")

	// Set deployer's balance to 10 ETH for contract deployment
	deployerBalance := C.CString("0x8ac7230489e80000") // 10 ETH in wei
	defer C.free(unsafe.Pointer(deployerBalance))

	if C.revm_set_balance(instance, deployer, deployerBalance) != 0 {
		fmt.Printf("❌ Failed to set deployer's balance: %s\n", getLastError(instance))
		return
	}
	fmt.Println("💰 Set deployer's balance to 10 ETH")

	// Deploy a simple contract
	fmt.Println("\n📦 Deploying contract...")
	
	// Simple contract bytecode that stores a value and returns it
	bytecode := []byte{
		0x60, 0x80, 0x60, 0x40, 0x52, // Set up memory
		0x34, 0x80, 0x15, 0x61, 0x00, 0x10, 0x57, 0x60, 0x00, 0x80, 0xfd, // Check value
		0x5b, // JUMPDEST
		0x60, 0x01, // PUSH1 0x01 (return true)
		0x60, 0x00, // PUSH1 0x00 (memory position)
		0x52, // MSTORE (store in memory)
		0x60, 0x20, // PUSH1 0x20 (32 bytes)
		0x60, 0x00, // PUSH1 0x00 (memory position)
		0xf3, // RETURN
	}

	deployResult := C.revm_deploy_contract(
		instance,
		deployer,
		(*C.uchar)(unsafe.Pointer(&bytecode[0])),
		C.uint(len(bytecode)),
		1000000, // gas limit
	)

	if deployResult == nil {
		fmt.Printf("❌ Contract deployment failed: %s\n", getLastError(instance))
		return
	}
	defer C.revm_free_deployment_result(deployResult)

	if deployResult.success != 1 {
		fmt.Println("❌ Contract deployment was not successful")
		return
	}

	contractAddress := C.GoString(deployResult.contract_address)
	fmt.Printf("✅ Contract deployed at: %s\n", contractAddress)
	fmt.Printf("⛽ Gas used: %d\n", deployResult.gas_used)

	// Perform a simple transfer from Alice to Bob
	fmt.Println("\n🔄 Performing transfer from Alice to Bob...")

	transferValue := C.CString("0x1bc16d674ec80000") // 2 ETH in wei
	defer C.free(unsafe.Pointer(transferValue))

	// Set transaction parameters for transfer
	if C.revm_set_tx(
		instance,
		alice,                    // caller
		bob,                      // to
		transferValue,            // value
		nil,                      // data
		0,                        // data_len
		21000,                    // gas_limit
		nil,                      // gas_price (use default)
		0,                        // nonce
	) != 0 {
		fmt.Printf("❌ Failed to set transaction: %s\n", getLastError(instance))
		return
	}

	// Execute the transfer
	execResult := C.revm_execute_commit(instance)
	if execResult == nil {
		fmt.Printf("❌ Transaction execution failed: %s\n", getLastError(instance))
		return
	}
	defer C.revm_free_execution_result(execResult)

	if execResult.success == 1 {
		fmt.Printf("✅ Transfer successful! Gas used: %d\n", execResult.gas_used)
	} else {
		fmt.Printf("❌ Transfer failed with status: %d\n", execResult.success)
	}

	// Check balances after transfer
	fmt.Println("\n💰 Checking balances after transfer...")

	aliceBalanceAfter := C.revm_get_balance(instance, alice)
	if aliceBalanceAfter != nil {
		defer C.revm_free_string(aliceBalanceAfter)
		fmt.Printf("💰 Alice's balance: %s\n", C.GoString(aliceBalanceAfter))
	}

	bobBalanceAfter := C.revm_get_balance(instance, bob)
	if bobBalanceAfter != nil {
		defer C.revm_free_string(bobBalanceAfter)
		fmt.Printf("💰 Bob's balance: %s\n", C.GoString(bobBalanceAfter))
	}

	// Demonstrate storage operations
	fmt.Println("\n🗄️ Demonstrating storage operations...")

	contractAddr := C.CString(contractAddress)
	storageSlot := C.CString("0x0")
	storageValue := C.CString("0x42")
	defer C.free(unsafe.Pointer(contractAddr))
	defer C.free(unsafe.Pointer(storageSlot))
	defer C.free(unsafe.Pointer(storageValue))

	// Set storage value
	if C.revm_set_storage(instance, contractAddr, storageSlot, storageValue) == 0 {
		fmt.Println("✅ Storage value set successfully")

		// Get storage value
		retrievedValue := C.revm_get_storage(instance, contractAddr, storageSlot)
		if retrievedValue != nil {
			defer C.revm_free_string(retrievedValue)
			fmt.Printf("📖 Retrieved storage value: %s\n", C.GoString(retrievedValue))
		}
	} else {
		fmt.Printf("❌ Failed to set storage: %s\n", getLastError(instance))
	}

	fmt.Println("\n🎉 REVM FFI Go example completed successfully!")
}

func getLastError(instance *C.RevmInstance) string {
	errorPtr := C.revm_get_last_error(instance)
	if errorPtr != nil {
		return C.GoString(errorPtr)
	}
	return "Unknown error"
} 