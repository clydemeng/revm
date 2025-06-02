use std::time::Instant;
use std::collections::HashMap;
use revm::{
    primitives::{TxKind, Address, Bytes, U256, hex, KECCAK_EMPTY},
    primitives::hardfork::SpecId,
    database::{CacheDB, EmptyDBTyped},
    handler::{MainBuilder, MainContext, MainnetEvm, ExecuteEvm},
    context::{Context, TxEnv, BlockEnv, CfgEnv},
    context_interface::{
        result::{ExecutionResult, Output},
        ContextTr,
    },
    state::AccountInfo,
    ExecuteCommitEvm,
};

// Contract bytecode - same as other benchmarks
const ERC20_WITH_MINT_BYTECODE: &str = "6080604052348015600e575f5ffd5b506105348061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061004a575f3560e01c806318160ddd1461004e57806340c10f191461006c57806370a0823114610088578063a9059cbb146100b8575b5f5ffd5b6100566100e8565b60405161006391906102b6565b60405180910390f35b61008660048036038101906100819190610357565b6100ee565b005b6100a2600480360381019061009d9190610395565b61015c565b6040516100af91906102b6565b60405180910390f35b6100d260048036038101906100cd9190610357565b610170565b6040516100df91906103da565b60405180910390f35b60015481565b8060015f8282546100ff9190610420565b92505081905550805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8282546101519190610420565b925050819055505050565b5f602052805f5260405f205f915090505481565b5f815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205410156101f0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101e7906104ad565b60405180910390fd5b815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461023b91906104cb565b92505081905550815f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461028d9190610420565b925050819055506001905092915050565b5f819050919050565b6102b08161029e565b82525050565b5f6020820190506102c95f8301846102a7565b92915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102fc826102d3565b9050919050565b61030c816102f2565b8114610316575f5ffd5b50565b5f8135905061032781610303565b92915050565b6103368161029e565b8114610340575f5ffd5b50565b5f813590506103518161032d565b92915050565b5f5f6040838503121561036d5761036c6102cf565b5b5f61037a85828601610319565b925050602061038b85828601610343565b9150509250929050565b5f602082840312156103aa576103a96102cf565b5b5f6103b784828501610319565b91505092915050565b5f8115159050919050565b6103d4816103c0565b82525050565b5f6020820190506103ed5f8301846103cb565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61042a8261029e565b91506104358361029e565b925082820190508082111561044d5761044c6103f3565b5b92915050565b5f82825260208201905092915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f610497601483610453565b91506104a282610463565b602082019050919050565b5f6020820190508181035f8301526104c48161048b565b9050919050565b5f6104d58261029e565b91506104e08361029e565b92508282039050818111156104f8576104f76103f3565b5b9291505056fea2646970667358221220cb19b4849bfce8663cd0287ac9f324dbeeb9b43d26a167e5a17d8452191f599d64736f6c634300081c0033";

// Test addresses - same as other benchmarks
const DEPLOYER_ADDRESS: &str = "0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6";
const ALICE_ADDRESS: &str = "0x8ba1f109551bD432803012645aac136c12345678";
const BOB_ADDRESS: &str = "0x1234567890123456789012345678901234567890";
const CHARLIE_ADDRESS: &str = "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd";
const VIEW_CALLER_ADDRESS: &str = "0x0000000000000000000000000000000000000001"; // Fresh address for view calls

// Constants - same as other benchmarks
const TOTAL_SUPPLY: &str = "1000000000000000000000000000"; // 1 billion tokens
const TRANSFER_AMOUNT: &str = "1000000000000000000"; // 1 token

// Nonce tracker
struct NonceTracker {
    nonces: HashMap<Address, u64>,
}

impl NonceTracker {
    fn new() -> Self {
        Self {
            nonces: HashMap::new(),
        }
    }
    
    fn get_and_increment(&mut self, address: Address) -> u64 {
        let nonce = self.nonces.get(&address).copied().unwrap_or(0);
        self.nonces.insert(address, nonce + 1);
        nonce
    }
}

fn main() {
    println!("🚀 Pure Rust REVM ERC20 Transfer Benchmark");
    println!("==========================================");

    println!("\n📊 Benchmark: 1000 transfers on BSC Testnet");
    run_benchmark();
}

fn run_benchmark() {
    let start_time = Instant::now();
    let mut nonce_tracker = NonceTracker::new();

    // 1. Create REVM instance with BSC Testnet configuration
    print!("   🔧 Creating REVM instance... ");
    let mut evm = create_evm_instance();
    println!("✅ (Chain ID: 97, Spec ID: Cancun)");

    // 2. Setup accounts with initial balances
    print!("   💰 Setting up accounts... ");
    let account_setup_time = Instant::now();
    setup_accounts(&mut evm);
    println!("✅ ({:?})", account_setup_time.elapsed());

    // 3. Deploy ERC20 token
    print!("   📄 Deploying ERC20 token... ");
    let deploy_time = Instant::now();
    
    let contract_address = deploy_contract(&mut evm, &mut nonce_tracker, DEPLOYER_ADDRESS, ERC20_WITH_MINT_BYTECODE);
    match contract_address {
        Some(addr) => {
            println!("✅ {} ({:?})", addr, deploy_time.elapsed());
        }
        None => {
            println!("❌ Failed to deploy contract");
            return;
        }
    }
    let contract_address = contract_address.unwrap();

    // 4. Mint tokens to deployer
    print!("   🪙 Minting tokens to deployer... ");
    let mint_time = Instant::now();
    
    if !mint_tokens(&mut evm, &mut nonce_tracker, DEPLOYER_ADDRESS, &contract_address, TOTAL_SUPPLY) {
        println!("❌ Failed to mint tokens");
        return;
    }
    
    let total_supply = get_total_supply(&mut evm, &mut nonce_tracker, &contract_address);
    let deployer_balance = get_token_balance(&mut evm, &mut nonce_tracker, DEPLOYER_ADDRESS, &contract_address);
    println!("✅ Total supply: {}, Deployer balance: {} ({:?})", 
        format_token_amount(&total_supply), 
        format_token_amount(&deployer_balance), 
        mint_time.elapsed());

    // 5. Transfer tokens to Alice
    print!("   🎯 Setting up token balances... ");
    let token_setup_time = Instant::now();
    
    if !transfer_tokens(&mut evm, &mut nonce_tracker, DEPLOYER_ADDRESS, ALICE_ADDRESS, &contract_address, TOTAL_SUPPLY) {
        println!("❌ Failed to transfer tokens to Alice");
        return;
    }
    
    let alice_balance = get_token_balance(&mut evm, &mut nonce_tracker, ALICE_ADDRESS, &contract_address);
    println!("✅ Alice balance: {} ({:?})", format_token_amount(&alice_balance), token_setup_time.elapsed());

    // 6. Perform benchmark transfers
    print!("   🚀 Performing 1000 transfers... ");
    let transfer_time = Instant::now();

    let success = perform_transfers(&mut evm, &mut nonce_tracker, &contract_address, 1000);
    let transfer_duration = transfer_time.elapsed();

    if !success {
        println!("❌ Transfers failed");
        return;
    }

    println!("✅ ({:?})", transfer_duration);

    // 7. Verify final balances
    print!("   ✅ Verifying balances... ");
    let verify_time = Instant::now();
    
    let final_alice_balance = get_token_balance(&mut evm, &mut nonce_tracker, ALICE_ADDRESS, &contract_address);
    let final_bob_balance = get_token_balance(&mut evm, &mut nonce_tracker, BOB_ADDRESS, &contract_address);
    let final_charlie_balance = get_token_balance(&mut evm, &mut nonce_tracker, CHARLIE_ADDRESS, &contract_address);
    
    println!("✅ Alice: {}, Bob: {}, Charlie: {} ({:?})", 
        format_token_amount(&final_alice_balance), 
        format_token_amount(&final_bob_balance), 
        format_token_amount(&final_charlie_balance),
        verify_time.elapsed());

    // 8. Print summary
    let total_time = start_time.elapsed();
    let transfers_per_second = 1000.0 / transfer_duration.as_secs_f64();
    let avg_per_transfer = transfer_duration / 1000;

    println!("   📈 Summary:");
    println!("      • Total time: {:?}", total_time);
    println!("      • Transfer time: {:?}", transfer_duration);
    println!("      • Transfers/second: {:.2}", transfers_per_second);
    println!("      • Average per transfer: {:.3}µs", avg_per_transfer.as_nanos() as f64 / 1000.0);
}

fn create_evm_instance() -> MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>> {
    let cache_db = CacheDB::new(EmptyDBTyped::default());
    
    let ctx = Context::mainnet()
        .with_db(cache_db)
        .modify_cfg_chained(|cfg| {
            cfg.chain_id = 97; // BSC Testnet
            cfg.spec = SpecId::CANCUN; // Use Cancun hardfork like other benchmarks
        })
        .modify_block_chained(|block| {
            block.number = U256::from(1);
            block.timestamp = U256::from(std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            block.gas_limit = 30_000_000u64;
            block.basefee = 1u64;
        });
    
    ctx.build_mainnet()
}

fn setup_accounts(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>) {
    let balance = U256::from_str_radix("56bc75e2d630e0000", 16).unwrap(); // 100 ETH
    let small_balance = U256::from_str_radix("2386f26fc10000", 16).unwrap(); // 0.01 ETH
    
    let accounts = [
        (DEPLOYER_ADDRESS, balance),
        (ALICE_ADDRESS, balance),
        (BOB_ADDRESS, small_balance),
        (CHARLIE_ADDRESS, small_balance),
        (VIEW_CALLER_ADDRESS, small_balance),
    ];

    for (addr_str, bal) in accounts {
        let addr = addr_str.parse::<Address>().unwrap();
        let account_info = AccountInfo {
            balance: bal,
            nonce: 0,
            code_hash: KECCAK_EMPTY,
            code: None,
        };
        // Insert account directly into the database
        evm.ctx.db().insert_account_info(addr, account_info);
    }
}

fn deploy_contract(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, deployer: &str, bytecode: &str) -> Option<Address> {
    let deployer_addr = deployer.parse::<Address>().unwrap();
    let code = hex::decode(bytecode).ok()?;
    
    let tx_env = TxEnv {
        caller: deployer_addr,
        kind: TxKind::Create,
        data: Bytes::from(code),
        value: U256::ZERO,
        gas_limit: 1_000_000,
        gas_price: 1u128,
        nonce: nonce_tracker.get_and_increment(deployer_addr),
        chain_id: Some(97), // BSC Testnet
        ..Default::default()
    };
    
    match evm.transact_commit(tx_env) {
        Ok(result) => {
            match result {
                ExecutionResult::Success { output, .. } => {
                    match output {
                        Output::Create(_, Some(address)) => {
                            println!("   ✅ Contract deployed at: {}", address);
                            Some(address)
                        }
                        _ => {
                            println!("   ❌ Contract deployment failed: No address returned");
                            None
                        }
                    }
                }
                ExecutionResult::Revert { output, .. } => {
                    println!("   ❌ Contract deployment reverted: {}", hex::encode(&output));
                    None
                }
                ExecutionResult::Halt { reason, .. } => {
                    println!("   ❌ Contract deployment halted: {:?}", reason);
                    None
                }
            }
        }
        Err(e) => {
            println!("   ❌ Contract deployment error: {:?}", e);
            None
        }
    }
}

fn mint_tokens(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, to: &str, contract: &Address, amount: &str) -> bool {
    let to_addr = to.parse::<Address>().unwrap();
    let amount_u256 = U256::from_str_radix(amount, 10).unwrap();
    
    // Encode mint(address,uint256) call
    // Function selector: 0x40c10f19
    let mut call_data = Vec::new();
    call_data.extend_from_slice(&hex::decode("40c10f19").unwrap()); // mint function selector
    call_data.extend_from_slice(&[0u8; 12]); // padding for address
    call_data.extend_from_slice(to_addr.as_slice()); // to address
    let amount_bytes = amount_u256.to_be_bytes::<32>();
    call_data.extend_from_slice(&amount_bytes); // amount
    
    let tx_env = TxEnv {
        caller: DEPLOYER_ADDRESS.parse().unwrap(),
        kind: TxKind::Call(*contract),
        data: Bytes::from(call_data),
        value: U256::ZERO,
        gas_limit: 100_000,
        gas_price: 1u128,
        nonce: nonce_tracker.get_and_increment(DEPLOYER_ADDRESS.parse().unwrap()),
        chain_id: Some(97), // BSC Testnet
        ..Default::default()
    };
    
    match evm.transact_commit(tx_env) {
        Ok(result) => {
            match result {
                ExecutionResult::Success { .. } => true,
                _ => false
            }
        }
        Err(_) => false
    }
}

fn transfer_tokens(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, from: &str, to: &str, contract: &Address, amount: &str) -> bool {
    let to_addr = to.parse::<Address>().unwrap();
    let amount_u256 = U256::from_str_radix(amount, 10).unwrap();
    
    // Encode transfer(address,uint256) call
    // Function selector: 0xa9059cbb
    let mut call_data = Vec::new();
    call_data.extend_from_slice(&hex::decode("a9059cbb").unwrap()); // transfer function selector
    call_data.extend_from_slice(&[0u8; 12]); // padding for address
    call_data.extend_from_slice(to_addr.as_slice()); // to address
    let amount_bytes = amount_u256.to_be_bytes::<32>();
    call_data.extend_from_slice(&amount_bytes); // amount
    
    let tx_env = TxEnv {
        caller: from.parse().unwrap(),
        kind: TxKind::Call(*contract),
        data: Bytes::from(call_data),
        value: U256::ZERO,
        gas_limit: 100_000,
        gas_price: 1u128,
        nonce: nonce_tracker.get_and_increment(from.parse().unwrap()),
        chain_id: Some(97), // BSC Testnet
        ..Default::default()
    };
    
    match evm.transact_commit(tx_env) {
        Ok(result) => {
            match result {
                ExecutionResult::Success { .. } => true,
                _ => false
            }
        }
        Err(_) => false
    }
}

fn get_total_supply(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, contract: &Address) -> U256 {
    // Encode totalSupply() call
    // Function selector: 0x18160ddd
    let call_data = hex::decode("18160ddd").unwrap();
    
    let view_caller = VIEW_CALLER_ADDRESS.parse().unwrap();
    let current_nonce = nonce_tracker.get_and_increment(view_caller);
    
    let tx_env = TxEnv {
        caller: view_caller,
        kind: TxKind::Call(*contract),
        data: Bytes::from(call_data),
        value: U256::ZERO,
        gas_limit: 100_000,
        gas_price: 1u128,
        nonce: current_nonce,
        chain_id: Some(97), // BSC Testnet
        ..Default::default()
    };
    
    match evm.transact(tx_env) {
        Ok(result) => {
            match result {
                ExecutionResult::Success { output, .. } => {
                    match output {
                        Output::Call(data) => {
                            if data.len() >= 32 {
                                U256::from_be_slice(&data[..32])
                            } else {
                                U256::ZERO
                            }
                        }
                        _ => U256::ZERO
                    }
                }
                _ => U256::ZERO
            }
        }
        Err(_) => U256::ZERO
    }
}

fn get_token_balance(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, account: &str, contract: &Address) -> U256 {
    let account_addr = account.parse::<Address>().unwrap();
    
    // Encode balanceOf(address) call
    // Function selector: 0x70a08231
    let mut call_data = Vec::new();
    call_data.extend_from_slice(&hex::decode("70a08231").unwrap()); // balanceOf function selector
    call_data.extend_from_slice(&[0u8; 12]); // padding for address
    call_data.extend_from_slice(account_addr.as_slice()); // account address
    
    let view_caller = VIEW_CALLER_ADDRESS.parse().unwrap();
    let current_nonce = nonce_tracker.get_and_increment(view_caller);
    
    let tx_env = TxEnv {
        caller: view_caller,
        kind: TxKind::Call(*contract),
        data: Bytes::from(call_data),
        value: U256::ZERO,
        gas_limit: 100_000,
        gas_price: 1u128,
        nonce: current_nonce,
        chain_id: Some(97), // BSC Testnet
        ..Default::default()
    };
    
    match evm.transact(tx_env) {
        Ok(result) => {
            match result {
                ExecutionResult::Success { output, .. } => {
                    match output {
                        Output::Call(data) => {
                            if data.len() >= 32 {
                                U256::from_be_slice(&data[..32])
                            } else {
                                U256::ZERO
                            }
                        }
                        _ => U256::ZERO
                    }
                }
                _ => U256::ZERO
            }
        }
        Err(_) => U256::ZERO
    }
}

fn perform_transfers(evm: &mut MainnetEvm<Context<BlockEnv, TxEnv, CfgEnv, CacheDB<EmptyDBTyped<std::convert::Infallible>>, revm::context::Journal<CacheDB<EmptyDBTyped<std::convert::Infallible>>>, ()>>, nonce_tracker: &mut NonceTracker, contract: &Address, count: usize) -> bool {
    for i in 0..count {
        // Alternate between Bob and Charlie as recipients (same as other benchmarks)
        let recipient = if i % 2 == 0 { BOB_ADDRESS } else { CHARLIE_ADDRESS };
        
        if !transfer_tokens(evm, nonce_tracker, ALICE_ADDRESS, recipient, contract, TRANSFER_AMOUNT) {
            println!("Transfer {} failed", i + 1);
            return false;
        }
    }
    true
}

fn format_token_amount(amount: &U256) -> String {
    let divisor = U256::from(10).pow(U256::from(18));
    let tokens = amount / divisor;
    tokens.to_string()
} 