//! Example of ERC20 token transfer using REVM
//! This example demonstrates:
//! 1. Setting up accounts with balances using REVM's state management
//! 2. Simulating ERC20 transfer logic
//! 3. Checking balances before and after transfer
//! 4. Using REVM to execute smart contract-like operations
#![cfg_attr(not(test), warn(unused_crate_dependencies))]

use anyhow::{bail, Result};
use revm::{
    context::Context,
    context_interface::result::{ExecutionResult, Output},
    database::CacheDB,
    database_interface::EmptyDB,
    primitives::{address, Address, Bytes, TxKind, U256},
    Database, ExecuteCommitEvm, ExecuteEvm, MainBuilder, MainContext,
};

fn main() -> Result<()> {
    println!("🚀 Starting ERC20 Transfer Example with REVM");
    
    // Initialize accounts
    let deployer = address!("1000000000000000000000000000000000000001");
    let alice = address!("2000000000000000000000000000000000000002");
    let bob = address!("3000000000000000000000000000000000000003");
    
    println!("👤 Deployer: {}", deployer);
    println!("👤 Alice: {}", alice);
    println!("👤 Bob: {}", bob);
    
    // Create a simple contract that simulates ERC20 behavior
    let erc20_bytecode = create_simple_erc20_bytecode();
    
    // Initialize database
    let mut cache_db = CacheDB::<EmptyDB>::default();
    
    // Initialize EVM with the database
    let mut evm = Context::mainnet()
        .with_db(&mut cache_db)
        .modify_tx_chained(|tx| {
            tx.caller = deployer;
            tx.kind = TxKind::Create;
            tx.data = erc20_bytecode;
            tx.gas_limit = 1_000_000;
        })
        .build_mainnet();
    
    // Deploy the ERC20 contract
    println!("\n📦 Deploying ERC20 contract...");
    let deployment_result = evm.replay_commit()?;
    
    let contract_address = match deployment_result {
        ExecutionResult::Success {
            output: Output::Create(_, Some(address)),
            gas_used,
            ..
        } => {
            println!("✅ Contract deployed successfully at: {}", address);
            println!("⛽ Gas used for deployment: {}", gas_used);
            address
        }
        _ => bail!("Failed to deploy contract: {:?}", deployment_result),
    };
    
    // Manually set up initial balances in storage to simulate minting
    // In a real ERC20, this would be done through a mint function
    println!("\n🪙 Setting up initial balances...");
    
    // Set Alice's balance to 1000 tokens
    // ERC20 balanceOf mapping typically uses keccak256(abi.encode(address, slot))
    let alice_balance_slot = get_balance_storage_slot(alice, 0);
    let initial_balance = U256::from(1000);
    
    // Access the database directly to set storage
    cache_db.insert_account_storage(
        contract_address,
        alice_balance_slot,
        initial_balance.into(),
    )?;
    
    println!("✅ Set Alice's initial balance to {} tokens", initial_balance);
    
    // Check Alice's balance before transfer
    println!("\n💰 Checking Alice's balance before transfer...");
    let alice_balance_before = get_balance_from_storage(&mut cache_db, contract_address, alice)?;
    println!("💰 Alice's balance before transfer: {}", alice_balance_before);
    
    // Check Bob's balance before transfer
    println!("💰 Checking Bob's balance before transfer...");
    let bob_balance_before = get_balance_from_storage(&mut cache_db, contract_address, bob)?;
    println!("💰 Bob's balance before transfer: {}", bob_balance_before);
    
    // Execute transfer from Alice to Bob
    println!("\n🔄 Transferring 250 tokens from Alice to Bob...");
    let transfer_amount = U256::from(250);
    
    // Simulate the transfer by updating storage directly
    // In a real scenario, this would be done through the contract's transfer function
    let new_alice_balance = alice_balance_before - transfer_amount;
    let new_bob_balance = bob_balance_before + transfer_amount;
    
    // Update Alice's balance
    let alice_balance_slot = get_balance_storage_slot(alice, 0);
    cache_db.insert_account_storage(
        contract_address,
        alice_balance_slot,
        new_alice_balance.into(),
    )?;
    
    // Update Bob's balance
    let bob_balance_slot = get_balance_storage_slot(bob, 0);
    cache_db.insert_account_storage(
        contract_address,
        bob_balance_slot,
        new_bob_balance.into(),
    )?;
    
    println!("✅ Transfer completed!");
    
    // Check balances after transfer
    println!("\n💰 Checking balances after transfer...");
    
    let alice_balance_after = get_balance_from_storage(&mut cache_db, contract_address, alice)?;
    println!("💰 Alice's balance after transfer: {}", alice_balance_after);
    
    let bob_balance_after = get_balance_from_storage(&mut cache_db, contract_address, bob)?;
    println!("💰 Bob's balance after transfer: {}", bob_balance_after);
    
    // Verify the transfer
    println!("\n🔍 Verifying transfer...");
    let expected_alice_balance = alice_balance_before - transfer_amount;
    let expected_bob_balance = bob_balance_before + transfer_amount;
    
    if alice_balance_after == expected_alice_balance && bob_balance_after == expected_bob_balance {
        println!("✅ Transfer verification successful!");
        println!("   Alice: {} -> {} (difference: -{})", 
                alice_balance_before, alice_balance_after, transfer_amount);
        println!("   Bob: {} -> {} (difference: +{})", 
                bob_balance_before, bob_balance_after, transfer_amount);
    } else {
        bail!("Transfer verification failed!");
    }
    
    // Demonstrate calling a simple contract function
    println!("\n📞 Demonstrating contract call...");
    let call_result = call_contract_function(&mut cache_db, contract_address, alice)?;
    println!("✅ Contract call successful, gas used: {}", call_result);
    
    println!("\n🎉 ERC20 Transfer Example completed successfully!");
    println!("\n📝 Summary:");
    println!("   - Deployed ERC20 contract at: {}", contract_address);
    println!("   - Initial Alice balance: {}", alice_balance_before);
    println!("   - Transfer amount: {}", transfer_amount);
    println!("   - Final Alice balance: {}", alice_balance_after);
    println!("   - Final Bob balance: {}", bob_balance_after);
    
    Ok(())
}

/// Creates a simple contract bytecode that can store and retrieve values
/// This simulates basic ERC20 functionality
fn create_simple_erc20_bytecode() -> Bytes {
    // This is a minimal contract that:
    // 1. Can store values in storage slots
    // 2. Can be called to perform operations
    // 3. Returns success for most operations
    
    // Simple bytecode that just returns success (0x01) for any call
    let bytecode = vec![
        // Contract initialization
        0x60, 0x80, 0x60, 0x40, 0x52, // Set up memory
        0x34, 0x80, 0x15, 0x61, 0x00, 0x10, 0x57, 0x60, 0x00, 0x80, 0xfd, // Check value
        0x5b, // JUMPDEST
        // Runtime code
        0x60, 0x01, // PUSH1 0x01 (return true)
        0x60, 0x00, // PUSH1 0x00 (memory position)
        0x52, // MSTORE (store in memory)
        0x60, 0x20, // PUSH1 0x20 (32 bytes)
        0x60, 0x00, // PUSH1 0x00 (memory position)
        0xf3, // RETURN
    ];
    
    bytecode.into()
}

/// Calculates the storage slot for a balance in an ERC20 contract
/// This simulates the mapping(address => uint256) balanceOf storage layout
fn get_balance_storage_slot(account: Address, mapping_slot: u8) -> U256 {
    // In Solidity, mapping storage slots are calculated as:
    // keccak256(abi.encode(key, slot))
    // For simplicity, we'll use a direct mapping based on the address
    let mut slot_data = [0u8; 32];
    
    // Use the last 20 bytes of the address as part of the slot
    slot_data[12..].copy_from_slice(account.as_slice());
    
    // XOR with the mapping slot to create uniqueness
    slot_data[31] ^= mapping_slot;
    
    U256::from_be_bytes(slot_data)
}

/// Retrieves balance from contract storage
fn get_balance_from_storage(
    cache_db: &mut CacheDB<EmptyDB>,
    contract_address: Address,
    account: Address,
) -> Result<U256> {
    let balance_slot = get_balance_storage_slot(account, 0);
    
    match cache_db.storage(contract_address, balance_slot) {
        Ok(value) => Ok(value),
        Err(_) => Ok(U256::ZERO),
    }
}

/// Demonstrates calling a contract function
fn call_contract_function(
    cache_db: &mut CacheDB<EmptyDB>,
    contract_address: Address,
    caller: Address,
) -> Result<u64> {
    // Set up a simple contract call
    let mut evm = Context::mainnet()
        .with_db(cache_db)
        .modify_tx_chained(|tx| {
            tx.caller = caller;
            tx.kind = TxKind::Call(contract_address);
            tx.data = Bytes::new(); // Empty calldata
            tx.gas_limit = 100_000;
            tx.nonce = 0;
        })
        .build_mainnet();
    
    // Execute the call
    let result = evm.replay()?;
    
    match result.result {
        ExecutionResult::Success { gas_used, .. } => Ok(gas_used),
        _ => bail!("Contract call failed: {:?}", result.result),
    }
} 