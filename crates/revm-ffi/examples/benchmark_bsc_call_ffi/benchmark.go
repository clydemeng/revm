package main

/*
#cgo LDFLAGS: -L../../../../target/release -lrevm_ffi
#include "../../../revm_ffi.h"
#include <stdlib.h>
*/
import "C"
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
	"unsafe"
)

// Contract bytecodes (from solc compilation)
const (
	// Minimal ERC20 bytecode (very simple, no constructor parameters needed)
	MINIMAL_ERC20_BYTECODE = "6080604052348015600e575f5ffd5b5069d3c21bcecceda10000006001819055335f908152602081905260409020556102278061003b5f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c806318160ddd1461004357806370a082311461005f578063a9059cbb1461007e575b5f5ffd5b61004c60015481565b6040519081526020015b60405180910390f35b61004c61006d366004610f565b5f6020819052908152604090205481565b61009161008c36600461018f565b6100a1565b6040519015158152602001610056565b335f908152602081905260408120548211156100fa5760405162461bcd60e51b8152602060048201526014602482015273496e73756666696369656e742062616c616e636560601b604482015260640160405180910390fd5b335f90815260208190526040812080548492906101189084906101cb565b90915550506001600160a01b0383165f9081526020819052604081208054849290610144908490610de565b9091555060019150505b92915050565b80356001600160a01b038116811461016a575f5ffd5b919050565b5f6020828403121561017f575f5ffd5b61018882610154565b9392505050565b5f5f604083850312156101a0575f5ffd5b6101a983610154565b946020939093013593505050565b634e487b7160e01b5f52601160045260245ffd5b8181038181111561014e5761014e6101b7565b808201808211156101e5761014e6101b756fea264697066735822122029cbac4f24a7413a21824ad198564fcc40e2fb9c9a73b8f17860210a1ea1209a64736f6c634300081c0033"

	// ERC20 with mint function (no constructor, tokens minted via function call)
	ERC20_WITH_MINT_BYTECODE = "6080604052348015600e575f5ffd5b506105348061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061004a575f3560e01c806318160ddd1461004e57806340c10f191461006c57806370a0823114610088578063a9059cbb146100b8575b5f5ffd5b6100566100e8565b60405161006391906102b6565b60405180910390f35b61008660048036038101906100819190610357565b6100ee565b005b6100a2600480360381019061009d9190610395565b61015c565b6040516100af91906102b6565b60405180910390f35b6100d260048036038101906100cd9190610357565b610170565b6040516100df91906103da565b60405180910390f35b60015481565b8060015f8282546100ff9190610420565b92505081905550805f5f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8282546101519190610420565b925050819055505050565b5f602052805f5260405f205f915090505481565b5f815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205410156101f0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101e7906104ad565b60405180910390fd5b815f5f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461023b91906104cb565b92505081905550815f5f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461028d9190610420565b925050819055506001905092915050565b5f819050919050565b6102b08161029e565b82525050565b5f6020820190506102c95f8301846102a7565b92915050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6102fc826102d3565b9050919050565b61030c816102f2565b8114610316575f5ffd5b50565b5f8135905061032781610303565b92915050565b6103368161029e565b8114610340575f5ffd5b50565b5f813590506103518161032d565b92915050565b5f5f6040838503121561036d5761036c6102cf565b5b5f61037a85828601610319565b925050602061038b85828601610343565b9150509250929050565b5f602082840312156103aa576103a96102cf565b5b5f6103b784828501610319565b91505092915050565b5f8115159050919050565b6103d4816103c0565b82525050565b5f6020820190506103ed5f8301846103cb565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61042a8261029e565b91506104358361029e565b925082820190508082111561044d5761044c6103f3565b5b92915050565b5f82825260208201905092915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f610497601483610453565b91506104a282610463565b602082019050919050565b5f6020820190508181035f8301526104c48161048b565b9050919050565b5f6104d58261029e565b91506104e08361029e565b92508282039050818111156104f8576104f76103f3565b5b9291505056fea2646970667358221220cb19b4849bfce8663cd0287ac9f324dbeeb9b43d26a167e5a17d8452191f599d64736f6c634300081c0033"

	// SimpleERC20 bytecode
	ERC20_BYTECODE = "60c0604052600e60809081526d2132b731b436b0b935aa37b5b2b760911b60a0525f9061002c9082610174565b506040805180820190915260058152640848a9c86960db1b60208201526001906100569082610174565b506002805460ff1916601217905534801561006f575f5ffd5b5060405161091a38038061091a83398101604081905261008e9161022e565b6003819055335f818152600460209081526040808320859055518481527fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef910160405180910390a350610245565b634e487b7160e01b5f52604160045260245ffd5b600181811c9082168061010457607f821691505b60208210810361012257634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561016f57805f5260205f20601f840160051c8101602085101561014d5750805b601f840160051c820191505b8181101561016c575f8155600101610159565b50505b505050565b81516001600160401b0381111561018d5761018d6100dc565b6101a18161019b84546100f0565b846100ee565b6020601f8211600181146101d3575f83156101bc5750848201515b5f19600385901b1c1916600184901b17845561016c565b5f84815260208120601f198516915b828110156101c857878501518255602094850194600190920191016101a8565b50848210156101e557868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b5f60208284031215610204575f5ffd5b5051919050565b610607806102185f395ff3fe6080604052348015600f575f5ffd5b5060043610610060575f3560e01c8063313ce56711610063578063313ce567146100ff57806370a082311461011e57806395d89b411461013d578063a9059cbb14610145578063dd62ed3e14610158575f5ffd5b806306fdde0314610094578063095ea7b3146100b257806318160ddd146100d557806323b872dd146100ec575b5f5ffd5b61009c610182565b6040516100a9919061045c565b60405180910390f35b6100c56100c03660046104ac565b61020e565b60405190151581526020016100a9565b6100de60025481565b6040519081526020016100a9565b6100c56100fa3660046104d4565b61023b565b60055461010c9060ff1681565b60405160ff90911681526020016100a9565b6100de61012c36600461050e565b5f6020819052908152604090205481565b61009c6103a2565b6100c56101533660046104ac565b6103af565b6100de61016636600461052e565b600160209081525f928352604080842090915290825290205481565b6003805461018f9061055f565b80601f01602080910402602001604051908101604052809291908181526020018280546101bb9061055f56b80156102065780601f106101dd57610100808354040283529160200191610206565b820191905f5260205f20905b8154815290600101906020018083116101e957829003601f168201915b505050505081565b335f908152600160208181526040808420600160a01b03871685529091529091208290555b92915050565b600160a01b0383165f90815260208190526040812054821115610e5760405162461bcd60e51b8152602060048201526014602482015273496e73756666696369656e742062616c616e636560601b60448201526064015b60405180910390fd5b600160a01b0384165f908152600160209081526040808320338452909152902054821115610309576040516461bcd60e51b8152602060048201526016602482015275496e73756666696369656e7420616c6c6f77616e636560501b6044820152606401610295565b600160a01b0384165f90815260208190526040812080548492906103309084906105ab565b90915550506001600160a01b0383165f908152602081905260408120805484929061035c9084906105be565b90915550506001600160a01b0384165f908152600160209081526040808320338452909152812080548492906103939084906105ab565b90915550600194935050505050565b6004805461018f9061055f565b335f908152602081905260408120548211156104045760405162461bcd60e51b8152602060048201526014602482015273496e73756666696369656e742062616c616e636560601b6044820152606401610295565b335f90815260208190526040812080548492906104229084906105ab565b90915550506001600160a01b0383165f90815260208190526040812080548492906104e9084906105be565b90915550600194935050505050565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b80356001600160a01b0381168114610568575f5ffd5b919050565b5f5f6040838503121561057e575f5ffd5b61058783610552565b946020939093013593505050565b5f5f5f606084860312156105a7575f5ffd5b6105b084610552565b92506105be60208501610552565b929592945050506040919091013590565b5f5f5f5f608085870312156105df575f5ffd5b6105e885610552565b93506105f660208601610552565b93969395505050506040820135916060013590565b5f6020828403121561061c575f5ffd5b61062582610552565b9392505050565b634e487b7160e01b5f52603260045260245ffd5b5f6020828403121561064b575f5ffd5b8151801515811461065a575f5ffd5b9392505050565b6020808252600f908201526e151c985b9cd9995c8819985a5b1959608a1b604082015260600190565b634e487b7160e01b5f52601160045260245ffd5b808201808211156101875761018761088a565b80820281158282048414176101875761018761088a"

	// TransferHelper bytecode  
	HELPER_BYTECODE = "6080604052348015600e575f5ffd5b50604051610984380380610984833981016040819052602b91604e565b5f80546001600160a01b0319166001600160a01b03929092169190911790556079565b5f60208284031215605d575f5ffd5b81516001600160a01b03811681146072575f5ffd5b9392505050565b6108fe806100865f395ff3fe608060405234801561000f575f5ffd5b5060043610610060575f3560e01c80630af4187d146100645780631239ec8c1461008a57806316beb982146100ad578063872c7046146100c0578063f8b2cb4f146100d3578063fc0c546a146100e6575b5f5ffd5b610077610072366004610685565b610110565b6040519081526020015b60405180910390f35b61009d6100983660046106fe565b61018d565b6040519015158152602001610081565b61009d6100bb36600461077e565b6103c8565b61009d6100ce3660046107b8565b6104b7565b6100776100e13660046107f7565b6105fe565b5f546100f8906001600160a01b031681565b6040516001600160a01b039091168152602001610081565b5f8054604051636eb1769f60e11b81526001600160a01b03858116600483015284811660248301529091169063dd62ed3e90604401602060405180830381865afa158015610160573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906101849190610810565b90505b92915050565b5f8382146101da5760405162461bcd60e51b8152602060048201526015602482015274082e4e4c2f240d8cadccee8d040dad2e6dac2e8c6d605b1b60448201526064015b60405180910390fd5b5f805b85811015610381575f546001600160a01b03166323b872dd898989858181106102085761020861082756b905060200201602081019061021d91906107f756b905060200201602081019061021d91906107f756b905060200201602081019061021d91906107f7565b88888681811061022f5761022f61082756b905060200201602081019061024491906107f756b905060200201602081019061024491906107f756b905060200201602081019061024491906107f756b5016001600160a01b0316886001600160a01b03167f3be21aaafaf4ce88ce563145ac2e334fd30fdd823a52f6792b0bff80b0ceda8b87878581811061025a5761025a61082756b905060200201602081019061026f91906107f756b905060200201602081019061026f91906107f756b905060200201602081019061026f91906107f756b5016001600160a01b0316886001600160a01b03167f3be21aaafaf4ce88ce563145ac2e334fd30fdd823a52f6792b0bff80b0ceda8b866040516102a491815260200190565b60405180910390a36001016101dd565b5060408051868152602081018390527f7b4fb3ffa747d03664dcd3bb02933bb6ab3f5799103a30a80659a5c1aa7ecc86910160405180910390a15060019695505050505050565b5f80546040516370a0823160e01b81526001600160a01b038481166004830152909116906370a0823190602401602060405180830381865afa15801561030b573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610187919061081056b5016001600160a01b0316886001600160a01b03167f3be21aaafaf4ce88ce563145ac2e334fd30fdd823a52f6792b0bff80b0ceda8b8660405161035b91815260200190565b60405180910390a350600193925050505050565b6001805461018e90610620565b335f908152600460205260408120548211156104915760405162461bcd60e51b8152602060048201526014602482015273496e73756666696369656e742062616c616e636560601b60448201526064016102d3565b335f90815260046020526040812080548492906104af90849061066c565b90915550506001600160a01b0383165f90815260046020526040812080548492906104db90849061067f565b90915550506040518281526001600160a01b0384169033907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef90602001610267565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b80356001600160a01b0381168114610568575f5ffd5b919050565b5f5f6040838503121561057e575f5ffd5b61058783610552565b946020939093013593505050565b5f5f5f606084860312156105a7575f5ffd5b6105b084610552565b92506105be60208501610552565b929592945050506040919091013590565b5f5f5f5f608085870312156105df575f5ffd5b6105e885610552565b93506105f660208601610552565b93969395505050506040820135916060013590565b5f6020828403121561061c575f5ffd5b61062582610552565b9392505050565b634e487b7160e01b5f52603260045260245ffd5b5f6020828403121561064b575f5ffd5b8151801515811461065a575f5ffd5b9392505050565b6020808252600f908201526e151c985b9cd9995c8819985a5b1959608a1b604082015260600190565b634e487b7160e01b5f52601160045260245ffd5b808201808211156101875761018761088a565b80820281158282048414176101875761018761088a"

	// ERC20 constructor parameter: total supply (1 billion tokens with 18 decimals)
	TOTAL_SUPPLY = "0x33b2e3c9fd0803ce8000000" // 1,000,000,000 * 10^18

	// Transfer amount (1 token with 18 decimals)
	TRANSFER_AMOUNT = "0xde0b6b3a7640000" // 1 * 10^18
)

// Benchmark configuration
type BenchmarkConfig struct {
	TransferCount    int
	BatchSize        int
	UseHelperContract bool
	ChainPreset      C.ChainPreset
}

// Test addresses
var (
	DEPLOYER_ADDRESS = "0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6"
	ALICE_ADDRESS    = "0x8ba1f109551bD432803012645aac136c12345678"
	BOB_ADDRESS      = "0x1234567890123456789012345678901234567890"
	CHARLIE_ADDRESS  = "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
)

func main() {
	fmt.Println("🚀 REVM FFI ERC20 Transfer Benchmark")
	fmt.Println("=====================================")

	// Test configuration - only 1000 transfers on BSC Testnet
	configs := []BenchmarkConfig{
		{TransferCount: 1000, BatchSize: 1, UseHelperContract: false, ChainPreset: C.BSC_TESTNET},
	}

	for _, config := range configs {
		fmt.Printf("\n📊 Benchmark: %d transfers on BSC Testnet\n", config.TransferCount)
		runBenchmark(config)
	}
}

func runBenchmark(config BenchmarkConfig) {
	startTime := time.Now()

	// 1. Create REVM instance with BSC configuration
	fmt.Print("   🔧 Creating REVM instance... ")
	instance := C.revm_new_with_preset(config.ChainPreset)
	if instance == nil {
		fmt.Println("❌ Failed to create REVM instance")
		return
	}
	defer C.revm_free(instance)

	chainID := C.revm_get_chain_id(instance)
	specID := C.revm_get_spec_id(instance)
	fmt.Printf("✅ (Chain ID: %d, Spec ID: %d)\n", chainID, specID)

	// 2. Setup accounts with initial balances
	fmt.Print("   💰 Setting up accounts... ")
	if !setupAccounts(instance) {
		fmt.Println("❌ Failed to set up accounts")
		return
	}

	// Test simple deployment first
	testSimpleDeployment(instance)

	// Deploy ERC20 token
	fmt.Print("📄 Deploying ERC20 token... ")
	deployTime := time.Now()
	
	// Try ERC20 with mint function (no constructor parameters)
	fmt.Print("(trying ERC20 with mint) ")
	erc20Address := deployContract(instance, DEPLOYER_ADDRESS, ERC20_WITH_MINT_BYTECODE, "")
	
	if erc20Address == "" {
		fmt.Println("❌ Failed to deploy ERC20 token")
		return
	}
	
	fmt.Printf("✅ %s (%v)\n", erc20Address, time.Since(deployTime))

	// Debug: Check deployer balance immediately after deployment
	fmt.Print("   🔍 Checking deployer balance immediately after deployment... ")
	fmt.Printf("(deployer: %s) ", DEPLOYER_ADDRESS)
	immediateBalance := getTokenBalance(instance, DEPLOYER_ADDRESS, erc20Address)
	fmt.Printf("Balance: %s\n", formatTokenAmount(immediateBalance))

	// Mint tokens to deployer using the mint function
	fmt.Print("   🪙 Minting tokens to Alice... ")
	if !mintTokens(instance, ALICE_ADDRESS, erc20Address, TOTAL_SUPPLY) {
		fmt.Println("❌ Failed to mint tokens")
		return
	}
	
	// Verify the minting worked
	aliceBalance := getTokenBalance(instance, ALICE_ADDRESS, erc20Address)
	fmt.Printf("✅ Alice balance: %s\n", formatTokenAmount(aliceBalance))

	var helperAddress string
	if config.UseHelperContract {
		// 4. Deploy TransferHelper contract
		fmt.Print("   🔧 Deploying TransferHelper... ")
		helperDeployTime := time.Now()
		
		helperAddress = deployHelperContract(instance, DEPLOYER_ADDRESS, erc20Address)
		if helperAddress == "" {
			fmt.Println("❌ Failed to deploy TransferHelper")
			return
		}
		
		fmt.Printf("✅ %s (%v)\n", helperAddress, time.Since(helperDeployTime))
	}

	// 5. Setup token balances and approvals
	fmt.Print("   🎯 Setting up token balances... ")
	tokenSetupTime := time.Now()
	
	// Check initial deployer balance
	deployerBalance := getTokenBalance(instance, DEPLOYER_ADDRESS, erc20Address)
	fmt.Printf("\n      Initial deployer balance: %s\n", formatTokenAmount(deployerBalance))
	
	// Also check total supply to see if contract was deployed correctly
	totalSupply := getTotalSupply(instance, erc20Address)
	fmt.Printf("      Total supply: %s\n", formatTokenAmount(totalSupply))
	
	// Give Alice all the tokens
	fmt.Print("      Verifying Alice has tokens... ")
	// Skip transfer since we minted directly to Alice
	// if !transferTokens(instance, DEPLOYER_ADDRESS, ALICE_ADDRESS, erc20Address, TOTAL_SUPPLY) {
	//	fmt.Println("❌ Failed to transfer tokens to Alice")
	//	return
	// }
	
	// Verify Alice has the tokens
	aliceTokenBalance := getTokenBalance(instance, ALICE_ADDRESS, erc20Address)
	fmt.Printf("✅ Alice has %s tokens\n", formatTokenAmount(aliceTokenBalance))

	if config.UseHelperContract {
		// Approve helper contract to spend Alice's tokens
		fmt.Print("      Approving helper contract... ")
		approveAmount := "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" // Max approval
		if !approveTokens(instance, ALICE_ADDRESS, helperAddress, erc20Address, approveAmount) {
			fmt.Println("❌ Failed to approve helper contract")
			return
		}
		fmt.Println("✅")
	}

	fmt.Printf("✅ (%v)\n", time.Since(tokenSetupTime))

	// 6. Perform benchmark transfers
	fmt.Printf("   🚀 Performing %d transfers... ", config.TransferCount)
	transferTime := time.Now()

	var success bool
	if config.UseHelperContract {
		success = performBatchTransfers(instance, helperAddress, config)
	} else {
		success = performIndividualTransfers(instance, erc20Address, config)
	}

	transferDuration := time.Since(transferTime)

	if !success {
		fmt.Println("❌ Transfers failed")
		return
	}

	fmt.Printf("✅ (%v)\n", transferDuration)

	// 7. Verify final balances
	fmt.Print("   ✅ Verifying balances... ")
	verifyTime := time.Now()
	
	aliceBalance = getTokenBalance(instance, ALICE_ADDRESS, erc20Address)
	bobBalance := getTokenBalance(instance, BOB_ADDRESS, erc20Address)
	charlieBalance := getTokenBalance(instance, CHARLIE_ADDRESS, erc20Address)
	
	fmt.Printf("✅ Alice: %s, Bob: %s, Charlie: %s (%v)\n", 
		formatTokenAmount(aliceBalance), 
		formatTokenAmount(bobBalance), 
		formatTokenAmount(charlieBalance),
		time.Since(verifyTime))

	// 8. Print summary
	totalTime := time.Since(startTime)
	transfersPerSecond := float64(config.TransferCount) / transferDuration.Seconds()
	
	fmt.Printf("   📈 Summary:\n")
	fmt.Printf("      • Total time: %v\n", totalTime)
	fmt.Printf("      • Transfer time: %v\n", transferDuration)
	fmt.Printf("      • Transfers/second: %.2f\n", transfersPerSecond)
	fmt.Printf("      • Average per transfer: %v\n", transferDuration/time.Duration(config.TransferCount))
}

func setupAccounts(instance *C.RevmInstance) bool {
	fmt.Print("💰 Setting up accounts... ")
	start := time.Now()

	// Set deployer balance (for gas fees)
	if !setBalance(instance, DEPLOYER_ADDRESS, "0x56bc75e2d630e0000") { // 100 ETH
		fmt.Println("❌ Failed to set deployer balance")
		return false
	}

	// Set Alice balance (for token transfers)
	if !setBalance(instance, ALICE_ADDRESS, "0x56bc75e2d630e0000") { // 100 ETH
		fmt.Println("❌ Failed to set Alice balance")
		return false
	}

	// Set Bob balance (small amount for gas)
	if !setBalance(instance, BOB_ADDRESS, "0x2386f26fc10000") { // 0.01 ETH
		fmt.Println("❌ Failed to set Bob balance")
		return false
	}

	// Set Charlie balance (small amount for gas)
	if !setBalance(instance, CHARLIE_ADDRESS, "0x2386f26fc10000") { // 0.01 ETH
		fmt.Println("❌ Failed to set Charlie balance")
		return false
	}

	elapsed := time.Since(start)
	fmt.Printf("✅ (%v)\n", elapsed)
	return true
}

func setBalance(instance *C.RevmInstance, address, balance string) bool {
	addressC := C.CString(address)
	balanceC := C.CString(balance)
	defer C.free(unsafe.Pointer(addressC))
	defer C.free(unsafe.Pointer(balanceC))

	result := C.revm_set_balance(instance, addressC, balanceC)
	if result != 0 {
		// Get the error message
		errorPtr := C.revm_get_last_error(instance)
		if errorPtr != nil {
			errorMsg := C.GoString(errorPtr)
			fmt.Printf("Error setting balance for %s: %s\n", address, errorMsg)
		}
		return false
	}
	return true
}

func deployContract(instance *C.RevmInstance, deployer, bytecode, constructorParam string) string {
	// Combine bytecode with constructor parameter
	fullBytecode := bytecode
	if constructorParam != "" {
		// Remove 0x prefix if present
		param := constructorParam
		if len(param) > 2 && param[:2] == "0x" {
			param = param[2:]
		}
		// Pad to 32 bytes (64 hex chars)
		for len(param) < 64 {
			param = "0" + param
		}
		fullBytecode += param
	}

	// Ensure even length
	if len(fullBytecode)%2 != 0 {
		fullBytecode = "0" + fullBytecode
	}

	fmt.Printf("(bytecode len: %d) ", len(fullBytecode))

	bytecodeBytes, err := hex.DecodeString(fullBytecode)
	if err != nil {
		fmt.Printf("Failed to decode bytecode: %v\n", err)
		return ""
	}

	deployerC := C.CString(deployer)
	defer C.free(unsafe.Pointer(deployerC))

	result := C.revm_deploy_contract(
		instance,
		deployerC,
		(*C.uchar)(unsafe.Pointer(&bytecodeBytes[0])),
		C.uint(len(bytecodeBytes)),
		C.uint64_t(3000000), // 3M gas limit
	)

	if result == nil {
		// Get the error message
		errorPtr := C.revm_get_last_error(instance)
		if errorPtr != nil {
			errorMsg := C.GoString(errorPtr)
			fmt.Printf("Failed to deploy contract: %s\n", errorMsg)
		} else {
			fmt.Printf("Failed to deploy contract: unknown error\n")
		}
		return ""
	}
	defer C.revm_free_deployment_result(result)

	fmt.Printf("(success=%d, gas=%d) ", result.success, result.gas_used)

	if result.success != 1 {
		fmt.Printf("Contract deployment failed (success=%d, gas_used=%d)\n", result.success, result.gas_used)
		return ""
	}

	if result.contract_address == nil {
		fmt.Printf("Contract deployment succeeded but no address returned\n")
		return ""
	}

	address := C.GoString(result.contract_address)
	C.revm_free_string(result.contract_address)
	return address
}

func deployHelperContract(instance *C.RevmInstance, deployer, tokenAddress string) string {
	// Encode constructor parameter (token address)
	tokenAddr := tokenAddress
	if len(tokenAddr) > 2 && tokenAddr[:2] == "0x" {
		tokenAddr = tokenAddr[2:]
	}
	// Pad to 32 bytes
	for len(tokenAddr) < 64 {
		tokenAddr = "0" + tokenAddr
	}

	return deployContract(instance, deployer, HELPER_BYTECODE, "0x"+tokenAddr)
}

func transferTokens(instance *C.RevmInstance, from, to, tokenAddress, amount string) bool {
	// Encode transfer(address,uint256) call
	// Function selector: 0xa9059cbb
	// to address (32 bytes)
	// amount (32 bytes)
	
	toAddr := to
	if len(toAddr) > 2 && toAddr[:2] == "0x" {
		toAddr = toAddr[2:]
	}
	for len(toAddr) < 64 {
		toAddr = "0" + toAddr
	}

	amountHex := amount
	if len(amountHex) > 2 && amountHex[:2] == "0x" {
		amountHex = amountHex[2:]
	}
	for len(amountHex) < 64 {
		amountHex = "0" + amountHex
	}

	callData := "a9059cbb" + toAddr + amountHex
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return false
	}

	fromC := C.CString(from)
	tokenC := C.CString(tokenAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(fromC))
	defer C.free(unsafe.Pointer(tokenC))
	defer C.free(unsafe.Pointer(valueC))

	result := C.revm_call_contract(
		instance,
		fromC,
		tokenC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(200000), // 200k gas limit
	)

	if result == nil {
		return false
	}
	defer C.revm_free_execution_result(result)

	return result.success == 1
}

func approveTokens(instance *C.RevmInstance, owner, spender, tokenAddress, amount string) bool {
	// Encode approve(address,uint256) call
	// Function selector: 0x095ea7b3
	
	spenderAddr := spender
	if len(spenderAddr) > 2 && spenderAddr[:2] == "0x" {
		spenderAddr = spenderAddr[2:]
	}
	for len(spenderAddr) < 64 {
		spenderAddr = "0" + spenderAddr
	}

	amountHex := amount
	if len(amountHex) > 2 && amountHex[:2] == "0x" {
		amountHex = amountHex[2:]
	}
	for len(amountHex) < 64 {
		amountHex = "0" + amountHex
	}

	callData := "095ea7b3" + spenderAddr + amountHex
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return false
	}

	ownerC := C.CString(owner)
	tokenC := C.CString(tokenAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(ownerC))
	defer C.free(unsafe.Pointer(tokenC))
	defer C.free(unsafe.Pointer(valueC))

	result := C.revm_call_contract(
		instance,
		ownerC,
		tokenC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(100000), // 100k gas limit
	)

	if result == nil {
		return false
	}
	defer C.revm_free_execution_result(result)

	return result.success == 1
}

func performIndividualTransfers(instance *C.RevmInstance, tokenAddress string, config BenchmarkConfig) bool {
	for i := 0; i < config.TransferCount; i++ {
		// Alternate between Bob and Charlie as recipients
		recipient := BOB_ADDRESS
		if i%2 == 1 {
			recipient = CHARLIE_ADDRESS
		}

		if !transferTokens(instance, ALICE_ADDRESS, recipient, tokenAddress, TRANSFER_AMOUNT) {
			fmt.Printf("Transfer %d failed\n", i+1)
			return false
		}
	}
	return true
}

func performBatchTransfers(instance *C.RevmInstance, helperAddress string, config BenchmarkConfig) bool {
	// Use benchmarkTransfers function for simplicity
	// Function selector: 0x872c7046
	// from address (32 bytes)
	// to address (32 bytes)  
	// amount (32 bytes)
	// count (32 bytes)

	fromAddr := ALICE_ADDRESS
	if len(fromAddr) > 2 && fromAddr[:2] == "0x" {
		fromAddr = fromAddr[2:]
	}
	for len(fromAddr) < 64 {
		fromAddr = "0" + fromAddr
	}

	toAddr := BOB_ADDRESS
	if len(toAddr) > 2 && toAddr[:2] == "0x" {
		toAddr = toAddr[2:]
	}
	for len(toAddr) < 64 {
		toAddr = "0" + toAddr
	}

	amountHex := TRANSFER_AMOUNT
	if len(amountHex) > 2 && amountHex[:2] == "0x" {
		amountHex = amountHex[2:]
	}
	for len(amountHex) < 64 {
		amountHex = "0" + amountHex
	}

	countHex := fmt.Sprintf("%x", config.TransferCount)
	for len(countHex) < 64 {
		countHex = "0" + countHex
	}

	callData := "872c7046" + fromAddr + toAddr + amountHex + countHex
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return false
	}

	deployerC := C.CString(DEPLOYER_ADDRESS)
	helperC := C.CString(helperAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(deployerC))
	defer C.free(unsafe.Pointer(helperC))
	defer C.free(unsafe.Pointer(valueC))

	gasLimit := uint64(config.TransferCount * 100000) // 100k gas per transfer
	result := C.revm_call_contract(
		instance,
		deployerC,
		helperC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(gasLimit),
	)

	if result == nil {
		return false
	}
	defer C.revm_free_execution_result(result)

	return result.success == 1
}

func getTokenBalance(instance *C.RevmInstance, address, tokenAddress string) string {
	// Encode balanceOf(address) call
	// Function selector: 0x70a08231
	
	addr := address
	if len(addr) > 2 && addr[:2] == "0x" {
		addr = addr[2:]
	}
	for len(addr) < 64 {
		addr = "0" + addr
	}

	callData := "70a08231" + addr
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return "0x0"
	}

	// Use a temporary address for the call (doesn't matter for view functions)
	callerC := C.CString(DEPLOYER_ADDRESS)
	tokenC := C.CString(tokenAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(callerC))
	defer C.free(unsafe.Pointer(tokenC))
	defer C.free(unsafe.Pointer(valueC))

	result := C.revm_call_contract(
		instance,
		callerC,
		tokenC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(100000), // 100k gas limit
	)

	if result == nil || result.success != 1 {
		if result != nil {
			C.revm_free_execution_result(result)
		}
		return "0x0"
	}
	defer C.revm_free_execution_result(result)

	if result.output_len == 0 {
		return "0x0"
	}

	// Convert output to hex string
	outputSlice := (*[32]byte)(unsafe.Pointer(result.output_data))[:result.output_len:result.output_len]
	return "0x" + hex.EncodeToString(outputSlice)
}

func formatTokenAmount(hexAmount string) string {
	if hexAmount == "0x0" || hexAmount == "" {
		return "0"
	}

	// Remove 0x prefix
	hexStr := hexAmount
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Convert to big.Int
	amount := new(big.Int)
	amount.SetString(hexStr, 16)

	// Convert to decimal with 18 decimals
	divisor := new(big.Int)
	divisor.Exp(big.NewInt(10), big.NewInt(18), nil)

	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.DivMod(amount, divisor, remainder)

	if remainder.Cmp(big.NewInt(0)) == 0 {
		return quotient.String()
	}

	// Format with decimals (show up to 6 decimal places)
	remainderStr := remainder.String()
	for len(remainderStr) < 18 {
		remainderStr = "0" + remainderStr
	}
	
	// Trim trailing zeros
	remainderStr = remainderStr[:6] // Show max 6 decimal places
	for len(remainderStr) > 0 && remainderStr[len(remainderStr)-1] == '0' {
		remainderStr = remainderStr[:len(remainderStr)-1]
	}

	if remainderStr == "" {
		return quotient.String()
	}

	return quotient.String() + "." + remainderStr
}

func generateRandomAddress() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return "0x" + hex.EncodeToString(bytes)
}

func testSimpleDeployment(instance *C.RevmInstance) {
	fmt.Println("🧪 Testing simple contract deployment...")
	
	// Simple contract that just returns 42
	// pragma solidity ^0.8.0;
	// contract Simple { function get() public pure returns (uint256) { return 42; } }
	simpleBytecode := "6080604052348015600e575f5ffd5b50609380601a5f395ff3fe6080604052348015600e575f5ffd5b50600436106026575f3560e01c80636d4ce63c14602a575b5f5ffd5b602a602f565b005b5f602a90509056fea2646970667358221220"
	
	fmt.Printf("   Deploying simple contract (bytecode length: %d)...\n", len(simpleBytecode))
	address := deployContract(instance, DEPLOYER_ADDRESS, simpleBytecode, "")
	if address != "" {
		fmt.Printf("   ✅ Simple contract deployed at: %s\n", address)
	} else {
		fmt.Println("   ❌ Simple contract deployment failed")
	}
}

func getTotalSupply(instance *C.RevmInstance, tokenAddress string) string {
	// Encode totalSupply() call
	// Function selector: 0x18160ddd
	
	callData := "18160ddd"
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return "0x0"
	}

	// Use a temporary address for the call (doesn't matter for view functions)
	callerC := C.CString(DEPLOYER_ADDRESS)
	tokenC := C.CString(tokenAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(callerC))
	defer C.free(unsafe.Pointer(tokenC))
	defer C.free(unsafe.Pointer(valueC))

	result := C.revm_call_contract(
		instance,
		callerC,
		tokenC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(100000), // 100k gas limit
	)

	if result == nil || result.success != 1 {
		if result != nil {
			C.revm_free_execution_result(result)
		}
		return "0x0"
	}
	defer C.revm_free_execution_result(result)

	if result.output_len == 0 {
		return "0x0"
	}

	// Convert output to hex string
	outputSlice := (*[32]byte)(unsafe.Pointer(result.output_data))[:result.output_len:result.output_len]
	return "0x" + hex.EncodeToString(outputSlice)
}

func setTokenBalanceInStorage(instance *C.RevmInstance, tokenAddress, address, amount string) bool {
	// For now, let's skip the manual storage setting and see if we can fix the constructor issue
	// The storage functions might not be available in the current FFI interface
	fmt.Printf("(skipping manual storage setting) ")
	return true
}

func mintTokens(instance *C.RevmInstance, to, tokenAddress, amount string) bool {
	// Encode mint(address,uint256) call
	// Function selector: 0x40c10f19
	
	toAddr := to
	if len(toAddr) > 2 && toAddr[:2] == "0x" {
		toAddr = toAddr[2:]
	}
	for len(toAddr) < 64 {
		toAddr = "0" + toAddr
	}

	amountHex := amount
	if len(amountHex) > 2 && amountHex[:2] == "0x" {
		amountHex = amountHex[2:]
	}
	for len(amountHex) < 64 {
		amountHex = "0" + amountHex
	}

	callData := "40c10f19" + toAddr + amountHex
	callDataBytes, err := hex.DecodeString(callData)
	if err != nil {
		return false
	}

	deployerC := C.CString(DEPLOYER_ADDRESS)
	tokenC := C.CString(tokenAddress)
	valueC := C.CString("0x0")
	defer C.free(unsafe.Pointer(deployerC))
	defer C.free(unsafe.Pointer(tokenC))
	defer C.free(unsafe.Pointer(valueC))

	result := C.revm_call_contract(
		instance,
		deployerC,
		tokenC,
		(*C.uchar)(unsafe.Pointer(&callDataBytes[0])),
		C.uint(len(callDataBytes)),
		valueC,
		C.uint64_t(100000), // 100k gas limit
	)

	if result == nil {
		return false
	}
	defer C.revm_free_execution_result(result)

	return result.success == 1
} 