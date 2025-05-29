//! FFI bindings for REVM (Rust Ethereum Virtual Machine)
//! 
//! This crate provides C-compatible FFI bindings for REVM, allowing other languages
//! like Go to interact with REVM through CGO.
//! 
//! # Safety
//! 
//! All FFI functions are marked as `unsafe` and require careful handling of memory
//! and pointer lifetimes. Callers must ensure proper cleanup of allocated resources.

#![cfg_attr(not(test), warn(unused_crate_dependencies))]

use std::ffi::CString;
use std::os::raw::{c_char, c_int, c_uint};
use std::ptr;

use revm::{
    context::Context,
    database::CacheDB,
    database_interface::EmptyDB,
    ExecuteCommitEvm, ExecuteEvm, MainBuilder, MainContext,
};

mod types;
mod utils;

pub use types::*;
pub use utils::*;

/// Initialize a new REVM instance
/// Returns a pointer to the EVM instance or null on failure
#[no_mangle]
pub unsafe extern "C" fn revm_new() -> *mut RevmInstance {
    let cache_db = CacheDB::<EmptyDB>::default();
    let ctx = Context::mainnet().with_db(cache_db);
    let evm = ctx.build_mainnet();
    
    let instance = Box::new(RevmInstance {
        evm,
        last_error: None,
    });
    
    Box::into_raw(instance)
}

/// Free a REVM instance
#[no_mangle]
pub unsafe extern "C" fn revm_free(instance: *mut RevmInstance) {
    if !instance.is_null() {
        let _ = Box::from_raw(instance);
    }
}

/// Set transaction parameters
#[no_mangle]
pub unsafe extern "C" fn revm_set_tx(
    instance: *mut RevmInstance,
    caller: *const c_char,
    to: *const c_char,
    value: *const c_char,
    data: *const u8,
    data_len: c_uint,
    gas_limit: c_uint,
    gas_price: *const c_char,
    nonce: c_uint,
) -> c_int {
    if instance.is_null() {
        return -1;
    }
    
    let instance = &mut *instance;
    
    // Clear any previous error
    instance.last_error = None;
    
    match set_transaction_params(instance, caller, to, value, data, data_len, gas_limit, gas_price, nonce) {
        Ok(()) => 0,
        Err(e) => {
            instance.last_error = Some(e.to_string());
            -1
        }
    }
}

/// Execute a transaction (without committing state changes)
#[no_mangle]
pub unsafe extern "C" fn revm_execute(instance: *mut RevmInstance) -> *mut ExecutionResultFFI {
    if instance.is_null() {
        return ptr::null_mut();
    }
    
    let instance = &mut *instance;
    
    match instance.evm.replay() {
        Ok(result) => {
            let ffi_result = convert_execution_result(result.result);
            Box::into_raw(Box::new(ffi_result))
        }
        Err(e) => {
            instance.last_error = Some(format!("Execution failed: {:?}", e));
            ptr::null_mut()
        }
    }
}

/// Execute and commit a transaction
#[no_mangle]
pub unsafe extern "C" fn revm_execute_commit(instance: *mut RevmInstance) -> *mut ExecutionResultFFI {
    if instance.is_null() {
        return ptr::null_mut();
    }
    
    let instance = &mut *instance;
    
    match instance.evm.replay_commit() {
        Ok(result) => {
            let ffi_result = convert_execution_result(result);
            Box::into_raw(Box::new(ffi_result))
        }
        Err(e) => {
            instance.last_error = Some(format!("Execution failed: {:?}", e));
            ptr::null_mut()
        }
    }
}

/// Deploy a contract
#[no_mangle]
pub unsafe extern "C" fn revm_deploy_contract(
    instance: *mut RevmInstance,
    deployer: *const c_char,
    bytecode: *const u8,
    bytecode_len: c_uint,
    gas_limit: c_uint,
) -> *mut DeploymentResultFFI {
    if instance.is_null() || bytecode.is_null() {
        return ptr::null_mut();
    }
    
    let instance = &mut *instance;
    
    match deploy_contract_impl(instance, deployer, bytecode, bytecode_len, gas_limit) {
        Ok(result) => Box::into_raw(Box::new(result)),
        Err(e) => {
            instance.last_error = Some(e.to_string());
            ptr::null_mut()
        }
    }
}

/// Get account balance
#[no_mangle]
pub unsafe extern "C" fn revm_get_balance(
    instance: *mut RevmInstance,
    address: *const c_char,
) -> *mut c_char {
    if instance.is_null() || address.is_null() {
        return ptr::null_mut();
    }
    
    let instance = &mut *instance;
    
    match get_balance_impl(instance, address) {
        Ok(balance_str) => {
            match CString::new(balance_str) {
                Ok(c_str) => c_str.into_raw(),
                Err(_) => ptr::null_mut(),
            }
        }
        Err(e) => {
            instance.last_error = Some(e.to_string());
            ptr::null_mut()
        }
    }
}

/// Set account balance
#[no_mangle]
pub unsafe extern "C" fn revm_set_balance(
    instance: *mut RevmInstance,
    address: *const c_char,
    balance: *const c_char,
) -> c_int {
    if instance.is_null() || address.is_null() || balance.is_null() {
        return -1;
    }
    
    let instance = &mut *instance;
    
    match set_balance_impl(instance, address, balance) {
        Ok(()) => 0,
        Err(e) => {
            instance.last_error = Some(e.to_string());
            -1
        }
    }
}

/// Get storage value
#[no_mangle]
pub unsafe extern "C" fn revm_get_storage(
    instance: *mut RevmInstance,
    address: *const c_char,
    slot: *const c_char,
) -> *mut c_char {
    if instance.is_null() || address.is_null() || slot.is_null() {
        return ptr::null_mut();
    }
    
    let instance = &mut *instance;
    
    match get_storage_impl(instance, address, slot) {
        Ok(value_str) => {
            match CString::new(value_str) {
                Ok(c_str) => c_str.into_raw(),
                Err(_) => ptr::null_mut(),
            }
        }
        Err(e) => {
            instance.last_error = Some(e.to_string());
            ptr::null_mut()
        }
    }
}

/// Set storage value
#[no_mangle]
pub unsafe extern "C" fn revm_set_storage(
    instance: *mut RevmInstance,
    address: *const c_char,
    slot: *const c_char,
    value: *const c_char,
) -> c_int {
    if instance.is_null() || address.is_null() || slot.is_null() || value.is_null() {
        return -1;
    }
    
    let instance = &mut *instance;
    
    match set_storage_impl(instance, address, slot, value) {
        Ok(()) => 0,
        Err(e) => {
            instance.last_error = Some(e.to_string());
            -1
        }
    }
}

/// Get the last error message
#[no_mangle]
pub unsafe extern "C" fn revm_get_last_error(instance: *mut RevmInstance) -> *const c_char {
    if instance.is_null() {
        return ptr::null();
    }
    
    let instance = &*instance;
    
    match &instance.last_error {
        Some(error) => error.as_ptr() as *const c_char,
        None => ptr::null(),
    }
}

/// Free a C string allocated by this library
#[no_mangle]
pub unsafe extern "C" fn revm_free_string(s: *mut c_char) {
    if !s.is_null() {
        let _ = CString::from_raw(s);
    }
}

/// Free an execution result
#[no_mangle]
pub unsafe extern "C" fn revm_free_execution_result(result: *mut ExecutionResultFFI) {
    if !result.is_null() {
        let _ = Box::from_raw(result);
    }
}

/// Free a deployment result
#[no_mangle]
pub unsafe extern "C" fn revm_free_deployment_result(result: *mut DeploymentResultFFI) {
    if !result.is_null() {
        let _ = Box::from_raw(result);
    }
} 