//! FFI-compatible types for REVM

use std::os::raw::{c_char, c_int, c_uint};
use revm::{
    database::CacheDB,
    database_interface::EmptyDB,
    handler::MainnetEvm,
};

/// Main REVM instance structure
#[repr(C)]
pub struct RevmInstance {
    pub evm: MainnetEvm<revm::Context<revm::context::BlockEnv, revm::context::TxEnv, revm::context::CfgEnv, CacheDB<EmptyDB>, revm::Journal<CacheDB<EmptyDB>>, ()>>,
    pub last_error: Option<String>,
}

/// FFI-compatible execution result
#[repr(C)]
pub struct ExecutionResultFFI {
    pub success: c_int,
    pub gas_used: c_uint,
    pub gas_refunded: c_uint,
    pub output_data: *mut u8,
    pub output_len: c_uint,
    pub logs_count: c_uint,
    pub logs: *mut LogFFI,
    pub created_address: *mut c_char, // Only for contract creation
}

/// FFI-compatible log structure
#[repr(C)]
pub struct LogFFI {
    pub address: *mut c_char,
    pub topics_count: c_uint,
    pub topics: *mut *mut c_char,
    pub data: *mut u8,
    pub data_len: c_uint,
}

/// FFI-compatible deployment result
#[repr(C)]
pub struct DeploymentResultFFI {
    pub success: c_int,
    pub contract_address: *mut c_char,
    pub gas_used: c_uint,
    pub gas_refunded: c_uint,
}

impl ExecutionResultFFI {
    // This function is not used - conversion is handled in utils.rs
}

impl LogFFI {
    fn from_revm_log(log: revm::primitives::Log) -> Self {
        let address_str = format!("{:?}", log.address);
        let address_ptr = match std::ffi::CString::new(address_str) {
            Ok(c_string) => c_string.into_raw(),
            Err(_) => std::ptr::null_mut(),
        };

        let topics_count = log.data.topics().len() as c_uint;
        let topics_ptr = if log.data.topics().is_empty() {
            std::ptr::null_mut()
        } else {
            let topic_strings: Vec<*mut c_char> = log
                .data
                .topics()
                .iter()
                .map(|topic| {
                    let topic_str = format!("{:?}", topic);
                    match std::ffi::CString::new(topic_str) {
                        Ok(c_string) => c_string.into_raw(),
                        Err(_) => std::ptr::null_mut(),
                    }
                })
                .collect();
            let boxed = topic_strings.into_boxed_slice();
            Box::into_raw(boxed) as *mut *mut c_char
        };

        let data = log.data.data.to_vec();
        let data_len = data.len() as c_uint;
        let data_ptr = if data.is_empty() {
            std::ptr::null_mut()
        } else {
            let boxed = data.into_boxed_slice();
            Box::into_raw(boxed) as *mut u8
        };

        LogFFI {
            address: address_ptr,
            topics_count,
            topics: topics_ptr,
            data: data_ptr,
            data_len,
        }
    }
} 