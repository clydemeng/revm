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
    pub fn from_revm_result(result: revm::context_interface::result::ResultAndState) -> Self {
        match result.result {
            revm::context_interface::result::ExecutionResult::Success {
                reason: _,
                gas_used,
                gas_refunded,
                logs,
                output,
            } => {
                let (output_data, output_len, created_address) = match output {
                    revm::context_interface::result::Output::Call(bytes) => {
                        let data = bytes.to_vec();
                        let len = data.len() as c_uint;
                        let ptr = if data.is_empty() {
                            std::ptr::null_mut()
                        } else {
                            let boxed = data.into_boxed_slice();
                            Box::into_raw(boxed) as *mut u8
                        };
                        (ptr, len, std::ptr::null_mut())
                    }
                    revm::context_interface::result::Output::Create(bytes, address) => {
                        let data = bytes.to_vec();
                        let len = data.len() as c_uint;
                        let ptr = if data.is_empty() {
                            std::ptr::null_mut()
                        } else {
                            let boxed = data.into_boxed_slice();
                            Box::into_raw(boxed) as *mut u8
                        };
                        
                        let addr_ptr = if let Some(addr) = address {
                            let addr_str = format!("{:?}", addr);
                            match std::ffi::CString::new(addr_str) {
                                Ok(c_string) => c_string.into_raw(),
                                Err(_) => std::ptr::null_mut(),
                            }
                        } else {
                            std::ptr::null_mut()
                        };
                        
                        (ptr, len, addr_ptr)
                    }
                };

                // Convert logs
                let logs_count = logs.len() as c_uint;
                let logs_ptr = if logs.is_empty() {
                    std::ptr::null_mut()
                } else {
                    let ffi_logs: Vec<LogFFI> = logs.into_iter().map(LogFFI::from_revm_log).collect();
                    let boxed = ffi_logs.into_boxed_slice();
                    Box::into_raw(boxed) as *mut LogFFI
                };

                ExecutionResultFFI {
                    success: 1,
                    gas_used: gas_used as c_uint,
                    gas_refunded: gas_refunded as c_uint,
                    output_data,
                    output_len,
                    logs_count,
                    logs: logs_ptr,
                    created_address,
                }
            }
            revm::context_interface::result::ExecutionResult::Revert { gas_used, output } => {
                let data = output.to_vec();
                let len = data.len() as c_uint;
                let ptr = if data.is_empty() {
                    std::ptr::null_mut()
                } else {
                    let boxed = data.into_boxed_slice();
                    Box::into_raw(boxed) as *mut u8
                };

                ExecutionResultFFI {
                    success: 0,
                    gas_used: gas_used as c_uint,
                    gas_refunded: 0,
                    output_data: ptr,
                    output_len: len,
                    logs_count: 0,
                    logs: std::ptr::null_mut(),
                    created_address: std::ptr::null_mut(),
                }
            }
            revm::context_interface::result::ExecutionResult::Halt { reason: _, gas_used } => ExecutionResultFFI {
                success: -1,
                gas_used: gas_used as c_uint,
                gas_refunded: 0,
                output_data: std::ptr::null_mut(),
                output_len: 0,
                logs_count: 0,
                logs: std::ptr::null_mut(),
                created_address: std::ptr::null_mut(),
            },
        }
    }
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