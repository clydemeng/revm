#ifndef REVM_FFI_H
#define REVM_FFI_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque pointer to REVM instance
typedef struct RevmInstance RevmInstance;

// FFI-compatible execution result
typedef struct {
    int success;           // 1 = success, 0 = revert, -1 = halt
    unsigned int gas_used;
    unsigned int gas_refunded;
    unsigned char* output_data;
    unsigned int output_len;
    unsigned int logs_count;
    struct LogFFI* logs;
    char* created_address;  // Only for contract creation
} ExecutionResultFFI;

// FFI-compatible log structure
typedef struct LogFFI {
    char* address;
    unsigned int topics_count;
    char** topics;
    unsigned char* data;
    unsigned int data_len;
} LogFFI;

// FFI-compatible deployment result
typedef struct {
    int success;
    char* contract_address;
    unsigned int gas_used;
    unsigned int gas_refunded;
} DeploymentResultFFI;

// Core REVM functions

/**
 * Initialize a new REVM instance
 * Returns a pointer to the EVM instance or NULL on failure
 */
RevmInstance* revm_new(void);

/**
 * Free a REVM instance
 */
void revm_free(RevmInstance* instance);

/**
 * Set transaction parameters
 * @param instance REVM instance
 * @param caller Caller address (hex string)
 * @param to Recipient address (hex string, NULL for contract creation)
 * @param value Transaction value (hex string, NULL for 0)
 * @param data Transaction data
 * @param data_len Length of transaction data
 * @param gas_limit Gas limit
 * @param gas_price Gas price (hex string, NULL for default)
 * @param nonce Transaction nonce
 * @return 0 on success, -1 on failure
 */
int revm_set_tx(
    RevmInstance* instance,
    const char* caller,
    const char* to,
    const char* value,
    const unsigned char* data,
    unsigned int data_len,
    unsigned int gas_limit,
    const char* gas_price,
    unsigned int nonce
);

/**
 * Execute a transaction (without committing state changes)
 * @param instance REVM instance
 * @return Execution result or NULL on failure
 */
ExecutionResultFFI* revm_execute(RevmInstance* instance);

/**
 * Execute and commit a transaction
 * @param instance REVM instance
 * @return Execution result or NULL on failure
 */
ExecutionResultFFI* revm_execute_commit(RevmInstance* instance);

/**
 * Deploy a contract
 * @param instance REVM instance
 * @param deployer Deployer address (hex string)
 * @param bytecode Contract bytecode
 * @param bytecode_len Length of bytecode
 * @param gas_limit Gas limit
 * @return Deployment result or NULL on failure
 */
DeploymentResultFFI* revm_deploy_contract(
    RevmInstance* instance,
    const char* deployer,
    const unsigned char* bytecode,
    unsigned int bytecode_len,
    unsigned int gas_limit
);

// Account and storage functions

/**
 * Get account balance
 * @param instance REVM instance
 * @param address Account address (hex string)
 * @return Balance as hex string or NULL on failure
 */
char* revm_get_balance(RevmInstance* instance, const char* address);

/**
 * Set account balance
 * @param instance REVM instance
 * @param address Account address (hex string)
 * @param balance Balance (hex string)
 * @return 0 on success, -1 on failure
 */
int revm_set_balance(RevmInstance* instance, const char* address, const char* balance);

/**
 * Get storage value
 * @param instance REVM instance
 * @param address Contract address (hex string)
 * @param slot Storage slot (hex string)
 * @return Storage value as hex string or NULL on failure
 */
char* revm_get_storage(RevmInstance* instance, const char* address, const char* slot);

/**
 * Set storage value
 * @param instance REVM instance
 * @param address Contract address (hex string)
 * @param slot Storage slot (hex string)
 * @param value Storage value (hex string)
 * @return 0 on success, -1 on failure
 */
int revm_set_storage(
    RevmInstance* instance,
    const char* address,
    const char* slot,
    const char* value
);

// Error handling

/**
 * Get the last error message
 * @param instance REVM instance
 * @return Error message or NULL if no error
 */
const char* revm_get_last_error(RevmInstance* instance);

// Memory management

/**
 * Free a C string allocated by this library
 */
void revm_free_string(char* s);

/**
 * Free an execution result
 */
void revm_free_execution_result(ExecutionResultFFI* result);

/**
 * Free a deployment result
 */
void revm_free_deployment_result(DeploymentResultFFI* result);

#ifdef __cplusplus
}
#endif

#endif // REVM_FFI_H 